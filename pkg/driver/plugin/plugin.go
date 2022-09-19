package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

const defaultRunnersCount = 16

type driverShim interface {
	Authorize(driver.AuthorizeInput) driver.AuthorizeOutput
	FieldResolve(driver.FieldResolveInput) driver.FieldResolveOutput
	InterfaceResolveType(driver.InterfaceResolveTypeInput) driver.InterfaceResolveTypeOutput
	ScalarParse(driver.ScalarParseInput) driver.ScalarParseOutput
	ScalarSerialize(driver.ScalarSerializeInput) driver.ScalarSerializeOutput
	UnionResolveType(driver.UnionResolveTypeInput) driver.UnionResolveTypeOutput
	Stream(driver.StreamInput) driver.StreamOutput
	Stdout(ctx context.Context, name string) error
	Stderr(ctx context.Context, name string) error
	SubscriptionConnection(driver.SubscriptionConnectionInput) driver.SubscriptionConnectionOutput
	SubscriptionListen(driver.SubscriptionListenInput) driver.SubscriptionListenOutput
}

type driverClient struct {
	driverShim
	plugin *Plugin
}

// Client interface used to establish connection with plugin
type Client interface {
	Client() (plugin.ClientProtocol, error)
	Kill()
}

// DefaultPluginClient creates a default plugin client
func DefaultPluginClient(cfg *plugin.ClientConfig) Client {
	return plugin.NewClient(cfg)
}

// NewPluginClient creates new client for plugin
var NewPluginClient = DefaultPluginClient

func (d driverClient) SetSecrets(in driver.SetSecretsInput) driver.SetSecretsOutput {
	return d.plugin.SetSecrets(in)
}

type pluginResponse struct {
	data interface{}
	err  error
}

type pluginPayload struct {
	data interface{}
	out  chan *pluginResponse
}

type pluginRunner chan *pluginPayload

func (r pluginRunner) do(p *Plugin, payload *pluginPayload) {
	defer func() {
		p.getRunner <- r
	}()
	dri, err := p.getDriver()
	if err != nil {
		go func() {
			payload.out <- &pluginResponse{
				err: err,
			}
		}()
		return
	}
	var resp interface{}
	switch data := payload.data.(type) {
	case driver.AuthorizeInput:
		resp = dri.Authorize(data)
	case driver.FieldResolveInput:
		resp = dri.FieldResolve(data)
	case driver.InterfaceResolveTypeInput:
		resp = dri.InterfaceResolveType(data)
	case driver.ScalarParseInput:
		resp = dri.ScalarParse(data)
	case driver.ScalarSerializeInput:
		resp = dri.ScalarSerialize(data)
	case driver.UnionResolveTypeInput:
		resp = dri.UnionResolveType(data)
	case driver.StreamInput:
		resp = dri.Stream(data)
	case driver.SubscriptionConnectionInput:
		resp = dri.SubscriptionConnection(data)
	case driver.SubscriptionListenInput:
		resp = dri.SubscriptionListen(data)
	default:
		err = errors.New("unknown input")
	}
	go func() {
		payload.out <- &pluginResponse{
			data: resp,
			err:  err,
		}
	}()
}

// Plugin implements Driver interface by running an executable available on local
// fs. All user defined operations will be forwarded to plugin through GRPC protocol.
type Plugin struct {
	cmd          string
	getRunner    chan pluginRunner
	runners      []pluginRunner
	client       Client
	runnersCount uint8
	lock         sync.RWMutex
	clilock      sync.RWMutex
	secrets      driver.Secrets
	cmdRef       *exec.Cmd
	done         chan struct{}
}

func (p *Plugin) getRunnersCount() uint8 {
	runnersCount := p.runnersCount
	if runnersCount == 0 {
		runnersCount = defaultRunnersCount
	}
	return runnersCount
}

func (p *Plugin) createRunners() {
	if p.runners != nil {
		return
	}
	runnersCount := p.getRunnersCount()
	p.runners = make([]pluginRunner, runnersCount)
	p.getRunner = make(chan pluginRunner, runnersCount)
	for i := uint8(0); i < runnersCount; i++ {
		runner := make(pluginRunner)
		go func() {
			for payload := range runner {
				runner.do(p, payload)
			}
		}()
		p.getRunner <- runner
		p.runners[i] = runner
	}
}

func (p *Plugin) getClient() (plugin.ClientProtocol, error) {
	p.clilock.RLock()
	defer p.clilock.RUnlock()
	return p.client.Client()
}

func (p *Plugin) getClientShim() (driverShim, error) {
	rpcClient, err := p.getClient()
	if err != nil {
		return nil, err
	}
	raw, err := rpcClient.Dispense("driver_grpc")
	if err != nil {
		return nil, err
	}
	driver, ok := raw.(driverShim)
	if !ok {
		return nil, errors.New("GRPC plugin does not implement driver")
	}
	return driver, nil
}

func (p *Plugin) getDriver() (driver.Driver, error) {
	driver, err := p.getClientShim()
	if err != nil {
		return nil, err
	}
	return driverClient{driver, p}, nil
}

// ExecCommand creates new plugin command
var ExecCommand = exec.Command

func (p *Plugin) getRunners() []pluginRunner {
	p.lock.RLock()
	runners := p.runners
	p.lock.RUnlock()
	return runners
}

func (p *Plugin) setClient(cmd *exec.Cmd) {
	p.clilock.Lock()
	defer p.clilock.Unlock()
	p.client = NewPluginClient(&plugin.ClientConfig{
		HandshakeConfig: p.handshake(),
		Plugins: map[string]plugin.Plugin{
			"driver_grpc": &GRPC{},
		},
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           NewLogger("plugin"),
	})
}

func (p *Plugin) setCmdRef(cmd *exec.Cmd) {
	p.clilock.Lock()
	defer p.clilock.Unlock()
	p.cmdRef = cmd
}

func (p *Plugin) createClient() error {
	cmd := ExecCommand(p.cmd)
	for k, v := range p.secrets {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	createProcGroup(cmd)
	p.setClient(cmd)
	d, err := p.getClientShim()
	if err != nil {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return err
	}
	p.setCmdRef(cmd)
	ctx := context.Background()
	go func() {
		if err := d.Stdout(ctx, "plugin."+filepath.Base(cmd.Path)); err != nil {
			klog.Error(err)
		}
	}()
	go func() {
		if err := d.Stderr(ctx, "plugin."+filepath.Base(cmd.Path)); err != nil {
			klog.Error(err)
		}
	}()
	return err
}

func (p *Plugin) start() error {
	if p.getRunners() == nil {
		p.lock.Lock()
		defer p.lock.Unlock()
		if p.runners == nil {
			err := p.createClient()
			if err != nil {
				return err
			}
			done := make(chan struct{})
			go func() {
				ticker := time.NewTicker(time.Second * 5)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						rpcClient, err := p.client.Client()
						if err == nil {
							err = rpcClient.Ping()
						}
						if err != nil {
							err = p.createClient()
							if err != nil {
								klog.Fatal(errors.Wrap(err, "plugin error, quitting: "))
								return
							}
						}
					case <-done:
						return
					}
				}
			}()
			p.done = done
			p.createRunners()
		}
	}
	return nil
}

func (p *Plugin) handshake() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "STUCCO_DRIVER_PLUGIN",
		MagicCookieValue: filepath.Base(p.cmd),
	}
}

func (p *Plugin) do(data interface{}) (interface{}, error) {
	if err := p.start(); err != nil {
		return nil, err
	}
	payload := pluginPayload{
		data: data,
		out:  make(chan *pluginResponse),
	}
	r := <-p.getRunner
	r <- &payload
	resp := <-payload.out
	return resp.data, resp.err
}

// SetSecrets sets user provided secrets for plugin using environment variables
func (p *Plugin) SetSecrets(in driver.SetSecretsInput) driver.SetSecretsOutput {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.client != nil {
		return driver.SetSecretsOutput{
			Error: &driver.Error{
				Message: "cannot change secrets on running client",
			},
		}
	}
	for k, sec := range in.Secrets {
		p.secrets[k] = sec
	}
	return driver.SetSecretsOutput{}
}

// FieldResolve uses plugin to resolve a field on type
func (p *Plugin) FieldResolve(in driver.FieldResolveInput) driver.FieldResolveOutput {
	resp, err := p.do(in)
	if err != nil {
		return driver.FieldResolveOutput{
			Error: &driver.Error{
				Message: err.Error(),
			},
		}
	}
	return resp.(driver.FieldResolveOutput)
}

// InterfaceResolveType uses plugin to find interface type for user input
func (p *Plugin) InterfaceResolveType(in driver.InterfaceResolveTypeInput) driver.InterfaceResolveTypeOutput {
	resp, err := p.do(in)
	if err != nil {
		return driver.InterfaceResolveTypeOutput{
			Error: &driver.Error{
				Message: err.Error(),
			},
		}
	}
	return resp.(driver.InterfaceResolveTypeOutput)
}

// ScalarParse uses plugin to parse scalar
func (p *Plugin) ScalarParse(in driver.ScalarParseInput) driver.ScalarParseOutput {
	resp, err := p.do(in)
	if err != nil {
		return driver.ScalarParseOutput{
			Error: &driver.Error{
				Message: err.Error(),
			},
		}
	}
	return resp.(driver.ScalarParseOutput)
}

// ScalarSerialize uses plugin to serialize scalar
func (p *Plugin) ScalarSerialize(in driver.ScalarSerializeInput) driver.ScalarSerializeOutput {
	resp, err := p.do(in)
	if err != nil {
		return driver.ScalarSerializeOutput{
			Error: &driver.Error{
				Message: err.Error(),
			},
		}
	}
	return resp.(driver.ScalarSerializeOutput)
}

// UnionResolveType uses plugin to find union type for user input
func (p *Plugin) UnionResolveType(in driver.UnionResolveTypeInput) driver.UnionResolveTypeOutput {
	resp, err := p.do(in)
	if err != nil {
		return driver.UnionResolveTypeOutput{
			Error: &driver.Error{
				Message: err.Error(),
			},
		}
	}
	return resp.(driver.UnionResolveTypeOutput)
}

// Stream data through grpc plugin
func (p *Plugin) Stream(in driver.StreamInput) driver.StreamOutput {
	resp, err := p.do(in)
	if err != nil {
		return driver.StreamOutput{
			Error: &driver.Error{
				Message: err.Error(),
			},
		}
	}
	return resp.(driver.StreamOutput)
}

// SubscriptionConnection creates connection with plugin
func (p *Plugin) SubscriptionConnection(in driver.SubscriptionConnectionInput) driver.SubscriptionConnectionOutput {
	resp, err := p.do(in)
	if err != nil {
		return driver.SubscriptionConnectionOutput{
			Error: &driver.Error{
				Message: err.Error(),
			},
		}
	}
	return resp.(driver.SubscriptionConnectionOutput)
}

// SubscriptionListen creates listen stream with plugin
func (p *Plugin) SubscriptionListen(in driver.SubscriptionListenInput) driver.SubscriptionListenOutput {
	resp, err := p.do(in)
	if err != nil {
		return driver.SubscriptionListenOutput{
			Error: &driver.Error{
				Message: err.Error(),
			},
		}
	}
	return resp.(driver.SubscriptionListenOutput)
}

// Authorize uses plugin to authorize request
func (p *Plugin) Authorize(in driver.AuthorizeInput) driver.AuthorizeOutput {
	resp, err := p.do(in)
	if err != nil {
		return driver.AuthorizeOutput{
			Error: &driver.Error{
				Message: err.Error(),
			},
		}
	}
	return resp.(driver.AuthorizeOutput)
}

// Close plugin and stop all runners
func (p *Plugin) Close() (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.client == nil {
		return
	}
	clean := make(chan struct{})
	go func() {
		defer close(clean)
		// wait for all runners
		for i := uint8(0); i < p.getRunnersCount(); i++ {
			close(<-p.getRunner)
		}
	}()
	t := time.NewTimer(10 * time.Second)
	select {
	case <-clean:
		t.Stop()
	case <-t.C:
		klog.Error("could not finish all tasks")
	}
	if p.done != nil {
		close(p.done)
	}
	clean = make(chan struct{})
	go func() {
		defer close(clean)
		p.client.Kill()
	}()
	t.Reset(5 * time.Second)
	select {
	case <-clean:
		t.Stop()
	case <-t.C:
		if err := killTree(p.cmdRef); err != nil {
			klog.Error("could not kill all processes in group")
		}
	}
	return
}

func checkFile(fn string) bool {
	if !strings.HasPrefix(filepath.Base(fn), "stucco-") {
		return false
	}
	st, err := os.Stat(fn)
	ok := err == nil && !st.IsDir() && isExecutable(st)
	if !ok && err != os.ErrNotExist {
		if err != nil {
			klog.Infof("ignoring %s: %v", fn, err)
		} else {
			klog.Infof("ignoring %s: not an executable file", fn)
		}
	}
	return ok
}

// ExecCommandContext used to check to create command for checking plugin config
var ExecCommandContext = exec.CommandContext

func checkPlugin(fn string) ([]driver.Config, error) {
	var cfgs []driver.Config
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := ExecCommandContext(
		ctx,
		fn,
		"config",
	)
	out, err := cmd.Output()
	if err != nil {
		if len(out) > 0 {
			klog.V(5).Infof("config command failed with output: %v", string(out))
		}
		return nil, errors.Wrap(err, "could not execute config")
	}
	err = json.Unmarshal(out, &cfgs)
	return cfgs, err
}

func cleanup(p []*Plugin) func() {
	return func() {
		wg := sync.WaitGroup{}
		for _, plug := range p {
			if p == nil {
				continue
			}
			wg.Add(1)
			go func(plug *Plugin) {
				defer wg.Done()
				if err := plug.Close(); err != nil {
					klog.Error(err)
				}
			}(plug)
		}
		wg.Wait()
	}
}

// Config is a plugin configuration
type Config struct {
	// defines how many concurrent clients for request can be open at once
	Runners uint8
	// Cmd is an executable path to plugin
	Cmd string
}

// NewPlugin creates new plugin ready to be used.
// Plugin must be closed after usage
func NewPlugin(cfg Config) *Plugin {
	return &Plugin{
		runnersCount: cfg.Runners,
		cmd:          cfg.Cmd,
		secrets:      driver.Secrets{},
	}
}

// LoadDriverPlugins searches environment PATH for files matching
// stucco-<plugin-name> executables and adds them as handlers for
// specific runtimes.
// Plugin must atleast be runnable with only binary name
// and if argument config is provided, plugin is expected to list
// supported runtimes in JSON and exit.
func LoadDriverPlugins(cfg Config) func() {
	plugins := []*Plugin{}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			path := filepath.Join(dir, f.Name())
			if !checkFile(path) {
				continue
			}
			cfgs, err := checkPlugin(path)
			if err != nil {
				klog.Infof("ignoring %s: %v", path, err)
				continue
			}
			plugCfg := cfg
			plugCfg.Cmd = path
			plug := NewPlugin(plugCfg)
			for _, cfg := range cfgs {
				driver.Register(cfg, plug)
			}
			plugins = append(plugins, plug)
		}
	}
	return cleanup(plugins)
}
