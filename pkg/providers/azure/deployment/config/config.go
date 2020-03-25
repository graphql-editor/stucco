package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"
)

const azureCLIPath = "AzureCLIPath"

// Option for building azure config
type Option interface {
	withOpt(opt *options)
}

// Subscription available from configuration
type Subscription struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// SubscriptionSource allows overiding subscription selection behaviour
type SubscriptionSource interface {
	Select(ctx context.Context, s []Subscription) (Subscription, error)
}

// select subscription by trying to match sub with ID or name from opts
// in case of failure returns the first subscription
type defaultSubscriptionSource struct {
	opts *options
}

func (d defaultSubscriptionSource) Select(ctx context.Context, subs []Subscription) (s Subscription, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	if len(subs) == 0 {
		err = errors.Errorf("no subscription found")
		return
	}
	s = subs[0]
	if len(subs) > 1 && d.opts != nil {
		for _, sub := range subs {
			if s.ID == d.opts.subscriptionID || s.Name == d.opts.subscriptionName {
				s = sub
				break
			}
		}
	}
	return
}

// DefaultSubscriptionSource used by package
var DefaultSubscriptionSource = defaultSubscriptionSource{}

type options struct {
	azPath             string
	subscriptionID     string
	subscriptionName   string
	subscriptionSource SubscriptionSource
}

// CLIPathOpt provides an ability to set custom azure cli path
type CLIPathOpt string

func (a CLIPathOpt) withOpt(opt *options) {
	opt.azPath = string(a)
}

// SubscriptionIDOpt allows overriding subscription by id from envrionment/client
type SubscriptionIDOpt string

func (s SubscriptionIDOpt) withOpt(opt *options) {
	opt.subscriptionID = string(s)
}

// SubscriptionNameOpt allows overriding subscription by name from envrionment/client
type SubscriptionNameOpt string

func (s SubscriptionNameOpt) withOpt(opt *options) {
	opt.subscriptionName = string(s)
}

type subscriptionSourceOpt struct {
	SubscriptionSource
}

// SubscriptionSourceOpt allows overriding default subscription selection behaviour
func SubscriptionSourceOpt(s SubscriptionSource) Option {
	return subscriptionSourceOpt{s}
}

func (s subscriptionSourceOpt) withOpt(opt *options) {
	if s.SubscriptionSource != nil {
		opt.subscriptionSource = s.SubscriptionSource
	}
}

// Config for azure resource management
type Config struct {
	SubscriptionID string
	Authorizer     autorest.Authorizer
}

func newConfigFromEnvironment(withMsi bool) (c Config, err error) {
	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return
	}
	c.SubscriptionID = settings.GetSubscriptionID()
	authorizer, _ := settings.GetAuthorizer()
	// check if environment variables that set authorizer are actually set
	// so that we don't get default MSI authorizer unless requested
	if settings.Values[auth.ClientSecret] != "" ||
		settings.Values[auth.CertificatePath] != "" ||
		(settings.Values[auth.Username] != "" && settings.Values[auth.Password] != "") ||
		withMsi {
		c.Authorizer = authorizer
	}
	return
}

// NewConfigFromEnvironment loads config from environment variables
func NewConfigFromEnvironment() (c Config, err error) {
	return newConfigFromEnvironment(true)
}

func querySubscriptionFromCLI(ctx context.Context, q SubscriptionSource) (s Subscription, err error) {
	// The default install paths are used to find Azure CLI. This is for security, so that any path in the calling program's Path environment is not used to execute Azure CLI.
	azureCLIDefaultPathWindows := fmt.Sprintf("%s\\Microsoft SDKs\\Azure\\CLI2\\wbin; %s\\Microsoft SDKs\\Azure\\CLI2\\wbin", os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles"))

	// Default path for non-Windows.
	const azureCLIDefaultPath = "/bin:/sbin:/usr/bin:/usr/local/bin"

	// Execute Azure CLI to get token
	var cliCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cliCmd = exec.CommandContext(ctx, fmt.Sprintf("%s\\system32\\cmd.exe", os.Getenv("windir")))
		cliCmd.Env = os.Environ()
		cliCmd.Env = append(cliCmd.Env, fmt.Sprintf("PATH=%s;%s", os.Getenv(azureCLIPath), azureCLIDefaultPathWindows))
		cliCmd.Args = append(cliCmd.Args, "/c", "az")
	} else {
		cliCmd = exec.Command("az")
		cliCmd.Env = os.Environ()
		cliCmd.Env = append(cliCmd.Env, fmt.Sprintf("PATH=%s:%s", os.Getenv(azureCLIPath), azureCLIDefaultPath))
	}
	cliCmd.Args = append(cliCmd.Args, "account", "list")

	var stderr bytes.Buffer
	cliCmd.Stderr = &stderr

	var output []byte
	output, err = cliCmd.Output()
	if err != nil {
		err = errors.Errorf("Invoking Azure CLI failed with the following error: %s", stderr.String())
		return
	}

	var subscriptions []Subscription
	err = json.Unmarshal(output, &subscriptions)
	if err != nil {
		return
	}
	s, err = q.Select(ctx, subscriptions)
	return
}

func newConfigFromCLI(ctx context.Context, q SubscriptionSource) (c Config, err error) {
	if q != nil {
		var sub Subscription
		sub, err = querySubscriptionFromCLI(ctx, q)
		c.SubscriptionID = sub.ID
	}
	if err == nil {
		c.Authorizer, err = auth.NewAuthorizerFromCLI()
	}
	return
}

// NewConfigFromCLIWithSource allows user to provide logic needed to choose
// subscription from results returned by CLI.
func NewConfigFromCLIWithSource(ctx context.Context, q SubscriptionSource) (c Config, err error) {
	return newConfigFromCLI(ctx, q)
}

// NewConfigFromCLI loads configuration from azure cli
// Similar to github.com/Azure/go-autorest/autorest/azure/auth/cli.GetTokenFromCLI
func NewConfigFromCLI(ctx context.Context) (c Config, err error) {
	return NewConfigFromCLIWithSource(ctx, DefaultSubscriptionSource)
}

func mergeConfig(dst *Config, src Config) {
	if dst.Authorizer == nil {
		dst.Authorizer = src.Authorizer
	}
	if dst.SubscriptionID == "" {
		dst.SubscriptionID = src.SubscriptionID
	}
}

// NewConfig creates new  with options
func NewConfig(ctx context.Context, opts ...Option) (c Config, err error) {
	var options options
	options.subscriptionSource = defaultSubscriptionSource{
		opts: &options,
	}
	for _, opt := range opts {
		opt.withOpt(&options)
	}
	c = Config{
		SubscriptionID: options.subscriptionID,
	}
	// ignore error here and just merge whatever was successful
	fromEnv, _ := newConfigFromEnvironment(false)
	mergeConfig(&c, fromEnv)
	querier := options.subscriptionSource
	if c.SubscriptionID != "" {
		querier = nil
	}
	if querier != nil || c.Authorizer == nil {
		var fromCLI Config
		if options.azPath != "" {
			oldVal := os.Getenv(azureCLIPath)
			os.Setenv(azureCLIPath, options.azPath)
			defer func() {
				switch oldVal {
				case "":
					os.Unsetenv(azureCLIPath)
				default:
					os.Setenv(azureCLIPath, oldVal)
				}
			}()
		}
		fromCLI, err = NewConfigFromCLIWithSource(ctx, querier)
		if err == nil {
			mergeConfig(&c, fromCLI)
		}
	}
	// load default msi authorizer
	if c.Authorizer == nil {
		fromEnv, _ := newConfigFromEnvironment(true)
		mergeConfig(&c, fromEnv)
	}
	return
}
