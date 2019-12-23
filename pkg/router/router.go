package router

import (
	"errors"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/parser"
	"github.com/graphql-go/graphql"
)

// Router dispatches defined functions to a driver that handles them
type Router struct {
	Interfaces map[string]InterfaceConfig // Interfaces is a map of FaaS function configs used in determining concrete type of an interface
	Resolvers  map[string]ResolverConfig  // Resolvers is a map of FaaS function configs used in resolution
	Scalars    map[string]ScalarConfig    // Scalars is a map of FaaS function configs used in parsing and serializing custom scalars
	Unions     map[string]UnionConfig     // Unions is a map of FaaS function configs used in determining concrete type of an union
	Schema     graphql.Schema             // Parsed schema
	Secrets    SecretsConfig              // Secrets is a map of secret references
}

func (r *Router) bindInterfaces(c *parser.Config) error {
	for k, i := range r.Interfaces {
		dri, err := r.getDriver(driver.Config{
			Provider: i.Environment.Provider,
			Runtime:  i.Environment.Runtime,
		})
		if err != nil {
			return err
		}
		r.setDriverSecrets(dri)
		c.Interfaces[k] = Dispatch{
			Driver:  dri,
			TypeMap: &r.Schema,
		}.InterfaceResolveType(i)
	}
	return nil
}

func (r *Router) bindResolvers(c *parser.Config) error {
	for k, rs := range r.Resolvers {
		dri, err := r.getDriver(driver.Config{
			Provider: rs.Environment.Provider,
			Runtime:  rs.Environment.Runtime,
		})
		if err != nil {
			return err
		}
		c.Resolvers[k] = Dispatch{dri, &r.Schema}.FieldResolve(rs)
	}
	return nil
}

func (r *Router) bindScalars(c *parser.Config) error {
	for k, s := range r.Scalars {
		dri, err := r.getDriver(driver.Config{
			Provider: s.Environment.Provider,
			Runtime:  s.Environment.Runtime,
		})
		if err != nil {
			return err
		}
		c.Scalars[k] = Dispatch{
			Driver:  dri,
			TypeMap: &r.Schema,
		}.ScalarFunctions(s)
	}
	return nil
}

func (r *Router) bindUnions(c *parser.Config) error {
	for k, u := range r.Unions {
		dri, err := r.getDriver(driver.Config{
			Provider: u.Environment.Provider,
			Runtime:  u.Environment.Runtime,
		})
		if err != nil {
			return err
		}
		c.Unions[k] = Dispatch{
			Driver:  dri,
			TypeMap: &r.Schema,
		}.UnionResolveType(u)
	}
	return nil
}

func (r *Router) parserConfig() (parser.Config, error) {
	c := parser.Config{
		Interfaces: make(map[string]graphql.ResolveTypeFn, len(r.Interfaces)),
		Resolvers:  make(map[string]graphql.FieldResolveFn, len(r.Resolvers)),
		Scalars:    make(map[string]parser.ScalarFunctions, len(r.Scalars)),
		Unions:     make(map[string]graphql.ResolveTypeFn, len(r.Unions)),
	}
	for _, f := range []func(c *parser.Config) error{
		r.bindInterfaces,
		r.bindResolvers,
		r.bindScalars,
		r.bindUnions,
	} {
		if err := f(&c); err != nil {
			return parser.Config{}, err
		}
	}
	return c, nil
}

func (r *Router) parseSchema(c Config) error {
	source, err := c.rawSchema()
	if err != nil {
		return err
	}
	pConfig, err := r.parserConfig()
	if err != nil {
		return err
	}
	p := parser.NewParser(pConfig)
	schema, err := p.Parse(source)
	if err != nil {
		return err
	}
	r.Schema = schema
	return nil
}

func (r *Router) setDriverSecrets(dri driver.Driver) error {
	var err error
	secrets := dri.SetSecrets(driver.SetSecretsInput{
		Secrets: r.Secrets.Secrets,
	})
	if err == nil && secrets.Error != nil {
		err = errors.New(secrets.Error.Message)
	}
	return err
}

func (r *Router) getDriver(cfg driver.Config) (dri driver.Driver, err error) {
	dri = driver.GetDriver(cfg)
	if dri == nil {
		err = errors.New("driver not found")
		return
	}
	err = r.setDriverSecrets(dri)
	return
}

func (r *Router) load(c Config) error {
	r.Secrets = c.Secrets
	for k, i := range c.Interfaces {
		i.Environment = newEnvironment(i.Environment, c.Environment)
		r.Interfaces[k] = i
	}
	for k, rs := range c.Resolvers {
		rs.Environment = newEnvironment(rs.Environment, c.Environment)
		r.Resolvers[k] = rs
	}
	for k, s := range c.Scalars {
		s.Environment = newEnvironment(s.Environment, c.Environment)
		r.Scalars[k] = s
	}
	for k, u := range c.Unions {
		u.Environment = newEnvironment(u.Environment, c.Environment)
		r.Unions[k] = u
	}
	return r.parseSchema(c)
}

// NewRouter creates new function router
func NewRouter(c Config) (Router, error) {
	c.Environment.Merge(DefaultEnvironment())
	r := Router{
		Interfaces: make(map[string]InterfaceConfig, len(c.Interfaces)),
		Resolvers:  make(map[string]ResolverConfig, len(c.Resolvers)),
		Scalars:    make(map[string]ScalarConfig, len(c.Scalars)),
		Unions:     make(map[string]UnionConfig, len(c.Unions)),
	}
	err := r.load(c)
	return r, err
}
