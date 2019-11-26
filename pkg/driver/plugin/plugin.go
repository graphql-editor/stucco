package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/hashicorp/go-plugin"
)

type driverShim interface {
	FieldResolve(driver.FieldResolveInput) (driver.FieldResolveOutput, error)
	InterfaceResolveType(driver.InterfaceResolveTypeInput) (driver.InterfaceResolveTypeOutput, error)
	ScalarParse(driver.ScalarParseInput) (driver.ScalarParseOutput, error)
	ScalarSerialize(driver.ScalarSerializeInput) (driver.ScalarSerializeOutput, error)
	UnionResolveType(driver.UnionResolveTypeInput) (driver.UnionResolveTypeOutput, error)
	Stream(driver.StreamInput) (driver.StreamOutput, error)
	Stdout(name string) error
	Stderr(name string) error
}

type driverClient struct {
	driverShim
	plugin *Plugin
}

func (d driverClient) SetSecrets(in driver.SetSecretsInput) (driver.SetSecretsOutput, error) {
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
	case driver.FieldResolveInput:
		resp, err = dri.FieldResolve(data)
	case driver.InterfaceResolveTypeInput:
		resp, err = dri.InterfaceResolveType(data)
	case driver.ScalarParseInput:
		resp, err = dri.ScalarParse(data)
	case driver.ScalarSerializeInput:
		resp, err = dri.ScalarSerialize(data)
	case driver.UnionResolveTypeInput:
		resp, err = dri.UnionResolveType(data)
	case driver.StreamInput:
		resp, err = dri.Stream(data)
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

type Plugin struct {
	cmd          string
	getRunner    chan pluginRunner
	runners      []pluginRunner
	client       *plugin.Client
	runnersCount uint8
	lock         sync.RWMutex
	secrets      driver.Secrets
}

func (p *Plugin) createRunners() {
	if p.runners != nil {
		return
	}
	p.runners = make([]pluginRunner, p.runnersCount)
	p.getRunner = make(chan pluginRunner, p.runnersCount)
	for i := uint8(0); i < p.runnersCount; i++ {
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

func (p *Plugin) getClientShim() (driverShim, error) {
	rpcClient, err := p.client.Client()
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

func (p *Plugin) start() error {
	p.lock.RLock()
	if p.runners == nil {
		p.lock.RUnlock()
		p.lock.Lock()
		defer p.lock.Unlock()
		if p.runners == nil {
			cmd := exec.Command(p.cmd)
			for k, v := range p.secrets {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
			}
			p.client = plugin.NewClient(&plugin.ClientConfig{
				HandshakeConfig: p.handshake(),
				Plugins: map[string]plugin.Plugin{
					"driver_grpc": &grpc.DriverGRPCPlugin{},
				},
				Cmd:              cmd,
				AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
				Logger:           newLogger("plugin"),
			})
			d, err := p.getClientShim()
			if err != nil {
				return err
			}
			if err := d.Stdout("plugin." + filepath.Base(cmd.Path)); err != nil {
				return err
			}
			if err := d.Stderr("plugin." + filepath.Base(cmd.Path)); err != nil {
				return err
			}
			p.createRunners()
		}
	} else {
		p.lock.RUnlock()
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

func (p *Plugin) SetSecrets(in driver.SetSecretsInput) (driver.SetSecretsOutput, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.client != nil {
		return driver.SetSecretsOutput{}, errors.New("cannot change secrets on running client")
	}
	for k, sec := range in.Secrets {
		p.secrets[k] = sec
	}
	return driver.SetSecretsOutput{}, nil
}
func (p *Plugin) FieldResolve(in driver.FieldResolveInput) (driver.FieldResolveOutput, error) {
	resp, err := p.do(in)
	if err != nil {
		return driver.FieldResolveOutput{}, err
	}
	return resp.(driver.FieldResolveOutput), nil
}
func (p *Plugin) InterfaceResolveType(in driver.InterfaceResolveTypeInput) (driver.InterfaceResolveTypeOutput, error) {
	resp, err := p.do(in)
	if err != nil {
		return driver.InterfaceResolveTypeOutput{}, err
	}
	return resp.(driver.InterfaceResolveTypeOutput), nil
}
func (p *Plugin) ScalarParse(in driver.ScalarParseInput) (driver.ScalarParseOutput, error) {
	resp, err := p.do(in)
	if err != nil {
		return driver.ScalarParseOutput{}, err
	}
	return resp.(driver.ScalarParseOutput), nil
}
func (p *Plugin) ScalarSerialize(in driver.ScalarSerializeInput) (driver.ScalarSerializeOutput, error) {
	resp, err := p.do(in)
	if err != nil {
		return driver.ScalarSerializeOutput{}, err
	}
	return resp.(driver.ScalarSerializeOutput), nil
}
func (p *Plugin) UnionResolveType(in driver.UnionResolveTypeInput) (driver.UnionResolveTypeOutput, error) {
	resp, err := p.do(in)
	if err != nil {
		return driver.UnionResolveTypeOutput{}, err
	}
	return resp.(driver.UnionResolveTypeOutput), nil
}
func (p *Plugin) Stream(in driver.StreamInput) (driver.StreamOutput, error) {
	resp, err := p.do(in)
	if err != nil {
		return driver.StreamOutput{}, err
	}
	return resp.(driver.StreamOutput), nil
}

func checkFile(fn string) bool {
	if fn == "" {
		return false
	}
	if !strings.HasPrefix(filepath.Base(fn), "stucco-") {
		return false
	}
	st, err := os.Stat(fn)
	if err != nil {
		return false
	}
	m := st.Mode()
	return !m.IsDir() && m&0111 != 0
}

func checkPlugin(fn string) ([]driver.Config, error) {
	var cfgs []driver.Config
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(
		ctx,
		fn,
		"config",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
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
				plug.lock.Lock()
				defer plug.lock.Unlock()
				if plug.client == nil {
					return
				}
				clean := make(chan struct{})
				go func() {
					// wait for all runners
					for i := uint8(0); i < plug.runnersCount; i++ {
						<-plug.getRunner
					}
					clean <- struct{}{}
				}()
				t := time.NewTimer(10 * time.Second)
				select {
				case <-clean:
					t.Stop()
				case <-t.C:
					fmt.Println("could not finish all tasks")
				}
				plug.client.Kill()
				wg.Done()
			}(plug)
		}
		wg.Wait()
	}
}

type Config struct {
	Runners uint8
}

func LoadDriverPlugins(cfg Config) func() {
	runnersCount := cfg.Runners
	if runnersCount == 0 {
		runnersCount = 64
	}
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
				continue
			}
			plug := &Plugin{
				cmd:          path,
				runnersCount: runnersCount,
				secrets:      driver.Secrets{},
			}
			for _, cfg := range cfgs {
				driver.Register(cfg, plug)
			}
			plugins = append(plugins, plug)
		}
	}
	return cleanup(plugins)
}
