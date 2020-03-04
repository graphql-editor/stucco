package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/graphql-editor/stucco/pkg/types"
)

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
	Environment Environment                `json:"environment"` // Environment is a default config of a router
	Interfaces  map[string]InterfaceConfig `json:"interfaces"`  // Interfaces is a map of FaaS function configs used in determining concrete type of an interface
	Resolvers   map[string]ResolverConfig  `json:"resolvers"`   // Resolvers is a map of FaaS function configs used in resolution
	Scalars     map[string]ScalarConfig    `json:"scalars"`     // Scalars is a map of FaaS function configs used in parsing and serializing custom scalars
	Schema      string                     `json:"schema"`      // String with GraphQL schema or an URL to the schema
	Unions      map[string]UnionConfig     `json:"unions"`      // Unions is a map of FaaS function configs used in determining concrete type of an union
	Secrets     SecretsConfig              `json:"secrets"`     // Secrets is a map of references to secrets
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
	strings.TrimPrefix(c.Schema, "file://")
	b, err := ioutil.ReadFile(c.Schema)
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
	switch {
	case strings.HasPrefix(c.Schema, "http://"), strings.HasPrefix(c.Schema, "https://"):
		return c.httpSchema()
	case c.Schema == "":
		c.Schema = "./schema.graphql"
		fallthrough
	case isFile(c.Schema):
		fmt.Fprintf(os.Stderr, "file schema\n")
		return c.fileSchema()
	}
	return c.Schema, nil
}
