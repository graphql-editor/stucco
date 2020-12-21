package router

import (
	"context"
	"errors"
	"io"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/parser"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
)

// Router dispatches defined functions to a driver that handles them
type Router struct {
	Interfaces    map[string]InterfaceConfig // Interfaces is a map of FaaS function configs used in determining concrete type of an interface
	Resolvers     map[string]ResolverConfig  // Resolvers is a map of FaaS function configs used in resolution
	Scalars       map[string]ScalarConfig    // Scalars is a map of FaaS function configs used in parsing and serializing custom scalars
	Unions        map[string]UnionConfig     // Unions is a map of FaaS function configs used in determining concrete type of an union
	Schema        graphql.Schema             // Parsed schema
	Secrets       SecretsConfig              // Secrets is a map of secret references
	Subscriptions SubscriptionConfig
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
		if err := r.setDriverSecrets(dri); err != nil {
			return err
		}
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
	if err := r.parseSchema(c); err != nil {
		return err
	}
	if r.Schema.SubscriptionType() != nil {
		env := newEnvironment(r.Subscriptions.Environment, c.Environment)
		r.Subscriptions.Environment = env
		dri, err := r.getDriver(driver.Config{
			Provider: env.Provider,
			Runtime:  env.Runtime,
		})
		if err != nil {
			return err
		}
		r.Schema.AddExtensions(&SubscribeExtension{
			router: r,
			dri:    dri,
		})
	}
	return nil
}

// NewRouter creates new function router
func NewRouter(c Config) (Router, error) {
	c.Environment.Merge(DefaultEnvironment())
	r := Router{
		Interfaces:    make(map[string]InterfaceConfig, len(c.Interfaces)),
		Resolvers:     make(map[string]ResolverConfig, len(c.Resolvers)),
		Scalars:       make(map[string]ScalarConfig, len(c.Scalars)),
		Unions:        make(map[string]UnionConfig, len(c.Unions)),
		Subscriptions: c.Subscriptions,
	}
	err := r.load(c)
	return r, err
}

// SubscribeContext contains information about subscription execution
type SubscribeContext struct {
	Query          string                          `json:"query,omitempty"`
	VariableValues map[string]interface{}          `json:"variableValues,omitempty"`
	OperationName  string                          `json:"operationName,omitempty"`
	IsSubscription bool                            `json:"-"`
	Reader         driver.SubscriptionListenReader `json:"-"`
	formattedErr   []gqlerrors.FormattedError
	err            error
}

type subscriptionExtensionKeyType string

const subscriptionExtensionKey subscriptionExtensionKeyType = "subscriptionExtensionKey"

// SubscribeExtension implements custom behaviour for extension types
// It basically passes through most of the parsing and validation process
// but short circuts the execution returning a custom result that should further be
// processed.
type SubscribeExtension struct {
	router *Router
	dri    driver.Driver
}

// Init implements graphql.Extension
func (s *SubscribeExtension) Init(ctx context.Context, p *graphql.Params) context.Context {
	return context.WithValue(ctx, subscriptionExtensionKey, &SubscribeContext{
		Query:          p.RequestString,
		VariableValues: p.VariableValues,
		OperationName:  p.OperationName,
	})
}

// Name of the subscription extension
func (s *SubscribeExtension) Name() string {
	switch s.router.Subscriptions.Kind {
	case ExternalSubscription:
		return "subscriptionExternal"
	case RedirectSubscription:
		return "subscriptionRedirect"
	case DefaultSubscription, BlockingSubscription:
		fallthrough
	default:
		return "subscriptionBlocking"
	}
}

// ParseDidStart implements graphql.Extension
func (s *SubscribeExtension) ParseDidStart(ctx context.Context) (context.Context, graphql.ParseFinishFunc) {
	return ctx, func(err error) {
		subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
		subCtx.err = err
	}
}

// ValidationDidStart implements graphql.Extension
func (s *SubscribeExtension) ValidationDidStart(ctx context.Context) (context.Context, graphql.ValidationFinishFunc) {
	return ctx, func(err []gqlerrors.FormattedError) {
		subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
		subCtx.formattedErr = err
	}
}

// ExecutionDidStart implements graphql.Extension
func (s *SubscribeExtension) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	return ctx, func(r *graphql.Result) {
		subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
		if subCtx.IsSubscription && subCtx.err == nil && len(subCtx.formattedErr) == 0 {
			r.Data = nil
			r.Errors = nil
		}
	}
}

// ResolveFieldDidStart implements graphql.Extension
// Hijacks the resolution of root subscription fields
func (s *SubscribeExtension) ResolveFieldDidStart(ctx context.Context, info *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	rawSubscription, ok := ctx.Value(RawSubscriptionKey).(bool)
	rawSubscription = ok && rawSubscription
	if !rawSubscription &&
		info.Path.Prev == nil &&
		s.router.Schema.SubscriptionType().Name() == info.ParentType.Name() {
		subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
		subCtx.IsSubscription = true
	}
	return ctx, func(v interface{}, err error) {}
}

// HasResult implements graphql.Extension
func (s *SubscribeExtension) HasResult(ctx context.Context) bool {
	subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
	return subCtx.IsSubscription
}

type nopCloserWriter struct {
	io.Writer
}

func (w nopCloserWriter) Close() error {
	return nil
}

// BlockingSubscriptionPayload returns data for blocking subscriptions to be handled by external protocol
type BlockingSubscriptionPayload struct {
	Context SubscribeContext
	Reader  driver.SubscriptionListenReader
}

func (s *SubscribeExtension) internalSubscription(ctx *SubscribeContext) driver.SubscriptionListenOutput {
	out := s.dri.SubscriptionListen(driver.SubscriptionListenInput{
		Function:       s.router.Subscriptions.Listen,
		Query:          ctx.Query,
		VariableValues: ctx.VariableValues,
		OperationName:  ctx.OperationName,
	})
	return out
}

func (s *SubscribeExtension) externalSubscription(ctx *SubscribeContext) driver.SubscriptionConnectionOutput {
	if s.router.Subscriptions.CreateConnection.Name == "" {
		panic("connection create function required for external subscription")
	}
	return s.dri.SubscriptionConnection(driver.SubscriptionConnectionInput{
		Function:       s.router.Subscriptions.CreateConnection,
		Query:          ctx.Query,
		VariableValues: ctx.VariableValues,
		OperationName:  ctx.OperationName,
	})
}

// GetResult implements graphql.Extension
func (s *SubscribeExtension) GetResult(ctx context.Context) interface{} {
	subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
	switch s.router.Subscriptions.Kind {
	case DefaultSubscription, BlockingSubscription:
		out := s.internalSubscription(subCtx)
		if out.Error != nil {
			subCtx.err = errors.New(out.Error.Message)
			panic(out.Error.Message)
		}
		return BlockingSubscriptionPayload{
			Context: *subCtx,
			Reader:  out.Reader,
		}
	case ExternalSubscription:
		out := s.externalSubscription(subCtx)
		if out.Error != nil {
			subCtx.err = errors.New(out.Error.Message)
			panic(out.Error.Message)
		}
		return out
	case RedirectSubscription:
		panic("redirect subscription not implemented yet")
	}
	panic("unsupported subscription kind")
}
