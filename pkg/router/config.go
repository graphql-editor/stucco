package router

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/graphql-editor/stucco/pkg/types"
)

// SchemaEnv is a name of environment variable that will be checked for schema if
// one is not provided
const SchemaEnv = "STUCCO_SCHEMA"

var (
	defaultEnvironment = Environment{
		Provider: "local",
		Runtime:  "nodejs",
	}
)

// SetDefaultEnvironment for router
func SetDefaultEnvironment(e Environment) {
	e.Merge(defaultEnvironment)
	defaultEnvironment = e
}

// DefaultEnvironment return default environment
func DefaultEnvironment() Environment {
	return defaultEnvironment
}

// Environment runtime environment for a function
type Environment struct {
	Provider string `json:"provider,omitempty"`
	Runtime  string `json:"runtime,omitempty"`
}

// Merge environments, original values have
// higher priority
func (e *Environment) Merge(src Environment) {
	if e.Provider == "" {
		e.Provider = src.Provider
	}
	if e.Runtime == "" {
		e.Runtime = src.Runtime
	}
}

func newEnvironment(base *Environment, defaultEnv Environment) *Environment {
	dst := new(Environment)
	if base != nil {
		*dst = *base
	}
	dst.Merge(defaultEnv)
	return dst
}

// ResolverConfig defines function configuration for field resolver
type ResolverConfig struct {
	Environment *Environment   `json:"environment,omitempty"`
	Resolve     types.Function `json:"resolve"`
}

// ScalarConfig defines parse and serialize function configurations for scalar
type ScalarConfig struct {
	Environment *Environment   `json:"environment,omitempty"`
	Parse       types.Function `json:"parse"`
	Serialize   types.Function `json:"serialize"`
}

// InterfaceConfig defines function configuration for interface type resolution
type InterfaceConfig struct {
	Environment *Environment   `json:"environment,omitempty"`
	ResolveType types.Function `json:"resolveType"`
}

// UnionConfig defines function configuration for union type resolution
type UnionConfig struct {
	Environment *Environment   `json:"environment,omitempty"`
	ResolveType types.Function `json:"resolveType"`
}

// SecretsConfig defines a secret configuration for router
type SecretsConfig struct {
	Secrets map[string]string `json:"secrets,omitempty"`
}

// Config is a router configuration mapping defined endpoints with thier runtime config
type Config struct {
	Environment         Environment                   `json:"environment"`         // Environment is a default config of a router
	Interfaces          map[string]InterfaceConfig    `json:"interfaces"`          // Interfaces is a map of FaaS function configs used in determining concrete type of an interface
	Resolvers           map[string]ResolverConfig     `json:"resolvers"`           // Resolvers is a map of FaaS function configs used in resolution
	Scalars             map[string]ScalarConfig       `json:"scalars"`             // Scalars is a map of FaaS function configs used in parsing and serializing custom scalars
	Schema              string                        `json:"schema"`              // String with GraphQL schema or an URL to the schema
	Unions              map[string]UnionConfig        `json:"unions"`              // Unions is a map of FaaS function configs used in determining concrete type of an union
	Secrets             SecretsConfig                 `json:"secrets"`             // Secrets is a map of references to secrets
	Subscriptions       SubscriptionConfig            `json:"subscriptions"`       // Configure subscription behaviour
	SubscriptionConfigs map[string]SubscriptionConfig `json:"subscriptionConfigs"` // Configure subscription behaviour per field
}

// AddResolver creates a new resolver mapping in config
func (c *Config) AddResolver(rsv, fn string) {
	if c.Resolvers == nil {
		c.Resolvers = make(map[string]ResolverConfig)
	}
	c.Resolvers[rsv] = ResolverConfig{
		Resolve: types.Function{
			Name: fn,
		},
	}
}

// AddInterface creates a new interface resolve type mapping in config
func (c *Config) AddInterface(intrf, fn string) {
	if c.Interfaces == nil {
		c.Interfaces = make(map[string]InterfaceConfig)
	}
	c.Interfaces[intrf] = InterfaceConfig{
		ResolveType: types.Function{
			Name: fn,
		},
	}
}

// AddScalar creates a new mapping for scalar parse and serialization
func (c *Config) AddScalar(sclr, parse string, serialize string) {
	if c.Scalars == nil {
		c.Scalars = make(map[string]ScalarConfig)
	}
	c.Scalars[sclr] = ScalarConfig{
		Parse: types.Function{
			Name: parse,
		},
		Serialize: types.Function{
			Name: serialize,
		},
	}
}

// AddUnion creates a new mapping for union resolve type
func (c *Config) AddUnion(union, fn string) {
	if c.Unions == nil {
		c.Unions = make(map[string]UnionConfig)
	}
	c.Unions[union] = UnionConfig{
		ResolveType: types.Function{
			Name: fn,
		},
	}
}

// AddSchema adds a schema source
func (c *Config) AddSchema(schema string) {
	c.Schema = schema
}

func (c Config) httpSchema() (string, error) {
	resp, err := http.Get(c.Schema)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c Config) fileSchema() (string, error) {
	b, err := ioutil.ReadFile(strings.TrimPrefix(c.Schema, "file://"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func isFile(s string) bool {
	st, err := os.Stat(s)
	return err == nil && st != nil
}

func (c Config) rawSchema() (string, error) {
	if env := os.Getenv(SchemaEnv); c.Schema == "" && env != "" {
		c.Schema = env
	}
	switch {
	case strings.HasPrefix(c.Schema, "http://"), strings.HasPrefix(c.Schema, "https://"):
		return c.httpSchema()
	case c.Schema == "":
		c.Schema = "./schema.graphql"
		fallthrough
	case isFile(c.Schema):
		return c.fileSchema()
	}
	return c.Schema, nil
}

// SubscriptionKind defines allowed types of subscription
type SubscriptionKind uint8

// UnmarshalJSON implements json.Unmarshaler
func (s *SubscriptionKind) UnmarshalJSON(b []byte) error {
	var sk string
	if err := json.Unmarshal(b, &sk); err != nil {
		return err
	}
	switch sk {
	case "blocking":
		*s = BlockingSubscription
	case "external":
		*s = ExternalSubscription
	case "redirect":
		*s = RedirectSubscription
	default:
		return errors.New("subscription kind must be of: blocking, external, redirect")
	}
	return nil
}

const (
	// DefaultSubscription defaults to internal
	DefaultSubscription SubscriptionKind = iota
	// BlockingSubscription subscription keeps alive subscription connection until client disconnects and streams
	BlockingSubscription
	// ExternalSubscription subscription returns a value from extension with connection payload to external service
	ExternalSubscription
	// RedirectSubscription returns 302 response with a redirect address being a user generated address
	RedirectSubscription
)

// SubscriptionConfig configures subscription handling for stucco
type SubscriptionConfig struct {
	Environment      *Environment     `json:"environment,omitempty"`
	Kind             SubscriptionKind `json:"kind,omitempty"`
	CreateConnection types.Function   `json:"createConnection,omitempty"`
	Listen           types.Function   `json:"listen,omitempty"`
}
