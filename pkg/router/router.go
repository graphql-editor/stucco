package router

import (
	"context"
	"errors"
	"io"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/parser"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/ast"
)

// Router dispatches defined functions to a driver that handles them
type Router struct {
	Interfaces          map[string]InterfaceConfig    // Interfaces is a map of FaaS function configs used in determining concrete type of an interface
	Resolvers           map[string]ResolverConfig     // Resolvers is a map of FaaS function configs used in resolution
	Scalars             map[string]ScalarConfig       // Scalars is a map of FaaS function configs used in parsing and serializing custom scalars
	Unions              map[string]UnionConfig        // Unions is a map of FaaS function configs used in determining concrete type of an union
	Schema              graphql.Schema                // Parsed schema
	Secrets             SecretsConfig                 // Secrets is a map of secret references
	Subscriptions       SubscriptionConfig            // global subscription config
	SubscriptionConfigs map[string]SubscriptionConfig // subscription config per subscription field
	MaxDepth            int                           // allow limiting max depth of GraphQL recursion
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
		c.Resolvers[k] = Dispatch{
			Driver:   dri,
			TypeMap:  &r.Schema,
			MaxDepth: r.MaxDepth,
		}.FieldResolve(rs)
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
		ext, err := newSubscriptionExtension(c.Subscriptions, r, dri)
		if err != nil {
			return err
		}
		r.SubscriptionConfigs = c.SubscriptionConfigs
		for k, v := range r.SubscriptionConfigs {
			ext.Exclude(k)
			fenv := newEnvironment(v.Environment, *env)
			ndri := dri
			if fenv.Provider != env.Provider || fenv.Runtime != env.Runtime {
				ndri, err = r.getDriver(driver.Config{
					Provider: fenv.Provider,
					Runtime:  fenv.Runtime,
				})
			}
			if err == nil {
				var fext subscribeExtension
				if v.Kind == DefaultSubscription {
					v.Kind = c.Subscriptions.Kind
				}
				if fext, err = newSubscriptionExtension(v, r, ndri); err == nil {
					fext.Include(k)
					r.Schema.AddExtensions(fext)
				}
			}
			if err != nil {
				return err
			}
		}
		r.Schema.AddExtensions(ext)
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
		MaxDepth:      c.MaxDepth,
	}
	err := r.load(c)
	return r, err
}

// SubscribeContext contains information about subscription execution
type SubscribeContext struct {
	Context             context.Context
	Query               string                          `json:"query,omitempty"`
	VariableValues      map[string]interface{}          `json:"variableValues,omitempty"`
	OperationName       string                          `json:"operationName,omitempty"`
	OperationDefinition *types.OperationDefinition      `json:"operationDefinition,omitempty"`
	IsSubscription      bool                            `json:"-"`
	Reader              driver.SubscriptionListenReader `json:"-"`
	info                *graphql.ResolveInfo            `json:"-"`
	resolvedTo          interface{}                     `json:"-"`
	formattedErr        []gqlerrors.FormattedError
	err                 error
}

type subscriptionExtensionKeyType string

const subscriptionExtensionKey subscriptionExtensionKeyType = "subscriptionExtensionKey"

// SubscribeExtension implements custom behaviour for extension types
// It basically passes through most of the parsing and validation process
// but short circuts the execution returning a custom result that should further be
// processed.
type SubscribeExtension struct {
	router  *Router
	dri     driver.Driver
	include []string
	exclude []string
}

// Include field in handling
func (s *SubscribeExtension) Include(f string) {
	s.include = append(s.include, f)
}

// Exclude field from handling
func (s *SubscribeExtension) Exclude(f string) {
	s.exclude = append(s.exclude, f)
}

// Init implements graphql.Extension
func (s *SubscribeExtension) Init(ctx context.Context, p *graphql.Params) context.Context {
	if ctx.Value(subscriptionExtensionKey) == nil {
		ctx = context.WithValue(ctx, subscriptionExtensionKey, &SubscribeContext{
			Query:          p.RequestString,
			VariableValues: p.VariableValues,
			OperationName:  p.OperationName,
			Context:        ctx,
		})
	}
	return ctx
}

func (s *SubscribeExtension) handle(ctx *SubscribeContext) bool {
	if ctx.err != nil || len(ctx.formattedErr) > 0 {
		return false
	}
	fn := s.fieldName(ctx)
	if len(s.include) == 0 {
		for _, v := range s.exclude {
			if fn == v {
				return false
			}
		}
		return true
	}
	for _, v := range s.include {
		if fn == v {
			return true
		}
	}
	return false
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
		if subCtx.IsSubscription && s.handle(subCtx) {
			r.Data = nil
			r.Errors = nil
			if len(subCtx.formattedErr) != 0 {
				r.Errors = subCtx.formattedErr
			} else if subCtx.err != nil {
				r.Errors = gqlerrors.FormatErrors(graphql.NewLocatedError(subCtx.err, []ast.Node{}))
			}
		}
	}
}

// ResolveFieldDidStart implements graphql.Extension
// Hijacks the resolution of root subscription fields
func (s *SubscribeExtension) ResolveFieldDidStart(ctx context.Context, info *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	rawSubscription, _ := ctx.Value(RawSubscriptionKey).(bool)
	if !rawSubscription && isSubscription(info) {
		subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
		subCtx.IsSubscription = true
		op, ok := info.Operation.(*ast.OperationDefinition)
		if ok && subCtx.OperationDefinition == nil && info.Operation != nil {
			subCtx.info = info
			if s.handle(subCtx) {
				subCtx.OperationDefinition = makeOperationDefinition(op, info.Fragments)
			}
		}
	}
	return ctx, func(v interface{}, err error) {}
}

// HasResult implements graphql.Extension
func (s *SubscribeExtension) HasResult(ctx context.Context) bool {
	subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
	return subCtx.IsSubscription && s.handle(subCtx)
}

func (s *SubscribeExtension) fieldName(ctx *SubscribeContext) string {
	op, ok := ctx.info.Operation.(*ast.OperationDefinition)
	if !ok {
		return ""
	}
	return op.SelectionSet.Selections[0].(*ast.Field).Name.Value
}

func (s *SubscribeExtension) subscriptionConfig(ctx *SubscribeContext) (SubscriptionConfig, error) {
	if cfg, ok := s.router.SubscriptionConfigs[s.fieldName(ctx)]; ok {
		if len(ctx.OperationDefinition.SelectionSet) > 1 {
			return SubscriptionConfig{}, errors.New("only one field on subscription supported with separate configs")
		}
		if cfg.Kind == DefaultSubscription {
			cfg.Kind = s.router.Subscriptions.Kind
		}
		return cfg, nil
	}
	return s.router.Subscriptions, nil
}

type nopCloserWriter struct {
	io.Writer
}

func (w nopCloserWriter) Close() error {
	return nil
}

// BlockingSubscriptionPayload returns data for blocking subscriptions to be handled by blocking protocol
type BlockingSubscriptionPayload struct {
	Context SubscribeContext
	Reader  driver.SubscriptionListenReader
}

// BlockingSubscriptionHandler can be implemented by field returned from root object on API to prepare connection data
// If SubscriptionConnection returns an error, that error will be returned. If it returns nil error and nil output value, then further execution is attempted. Otherwise value returned by handler is used to prepare extension output.
type BlockingSubscriptionHandler interface {
	SubscriptionListen(driver.SubscriptionListenInput) (*driver.SubscriptionListenOutput, error)
}

// BlockingSubscriptionExtension is a blocking subscription extension
type BlockingSubscriptionExtension struct {
	SubscribeExtension
}

// Name implements graphql.Extension
func (b *BlockingSubscriptionExtension) Name() string {
	return "subscriptionBlocking"
}

func (b *BlockingSubscriptionExtension) kind() SubscriptionKind {
	return BlockingSubscription
}

func (b *BlockingSubscriptionExtension) internalSubscription(ctx *SubscribeContext) (driver.SubscriptionListenOutput, error) {
	dri := b.dri
	cfg, err := b.subscriptionConfig(ctx)
	var out driver.SubscriptionListenOutput
	if err == nil {
		in := driver.SubscriptionListenInput{
			Function:       cfg.Listen,
			Query:          ctx.Query,
			VariableValues: ctx.VariableValues,
			OperationName:  ctx.OperationName,
			Operation:      ctx.OperationDefinition,
			Protocol:       ctx.Context.Value(ProtocolKey),
		}
		var nout *driver.SubscriptionListenOutput
		if h, ok := ctx.resolvedTo.(BlockingSubscriptionHandler); ok {
			nout, err = h.SubscriptionListen(in)
		}
		if err == nil {
			if nout != nil {
				out = *nout
			} else {
				out = dri.SubscriptionListen(in)
			}
		}
	}
	return out, err
}

// GetResult implements graphql.Extension
func (b *BlockingSubscriptionExtension) GetResult(ctx context.Context) interface{} {
	subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
	tout, err := b.internalSubscription(subCtx)
	if err != nil || tout.Error != nil {
		if err == nil {
			err = errors.New(tout.Error.Message)
		}
		panic(err)
	}
	return BlockingSubscriptionPayload{
		Reader:  tout.Reader,
		Context: *subCtx,
	}
}

// ExternalSubscriptionHandler can be implemented by root object on API to prepare connection data
// If SubscriptionConnection returns an error, that error will be returned. If it returns nil error and nil output value, then further execution is attempted. Otherwise value returned by handler is used to prepare extension output.
type ExternalSubscriptionHandler interface {
	SubscriptionConnection(driver.SubscriptionConnectionInput) (*driver.SubscriptionConnectionOutput, error)
}

// ExternalSubscriptionExtension is a blocking subscription extension
type ExternalSubscriptionExtension struct {
	SubscribeExtension
}

// Name implements graphql.Extension
func (e *ExternalSubscriptionExtension) Name() string {
	return "subscriptionExternal"
}

func (e *ExternalSubscriptionExtension) kind() SubscriptionKind {
	return ExternalSubscription
}

func (e *ExternalSubscriptionExtension) externalSubscription(ctx *SubscribeContext) (driver.SubscriptionConnectionOutput, error) {
	dri := e.dri
	cfg, err := e.subscriptionConfig(ctx)
	if err == nil && cfg.CreateConnection.Name == "" {
		err = errors.New("connection create function required for external subscription")
	}
	var out driver.SubscriptionConnectionOutput
	if err == nil {
		in := driver.SubscriptionConnectionInput{
			Function:       cfg.CreateConnection,
			Query:          ctx.Query,
			VariableValues: ctx.VariableValues,
			OperationName:  ctx.OperationName,
			Protocol:       ctx.Context.Value(ProtocolKey),
		}
		var nout *driver.SubscriptionConnectionOutput
		if h, ok := ctx.resolvedTo.(ExternalSubscriptionHandler); ok {
			nout, err = h.SubscriptionConnection(in)
		}
		if err == nil {
			if nout != nil {
				out = *nout
			} else {
				out = dri.SubscriptionConnection(in)
			}
		}
	}
	return out, err
}

// GetResult implements graphql.Extension
func (e *ExternalSubscriptionExtension) GetResult(ctx context.Context) interface{} {
	subCtx := ctx.Value(subscriptionExtensionKey).(*SubscribeContext)
	tout, err := e.externalSubscription(subCtx)
	if err != nil || tout.Error != nil {
		if err == nil {
			err = errors.New(tout.Error.Message)
		}
		panic(err)
	}
	return tout
}

type subscribeExtension interface {
	graphql.Extension
	Exclude(string)
	Include(string)
}

func newSubscriptionExtension(cfg SubscriptionConfig, r *Router, dri driver.Driver) (subscribeExtension, error) {
	ext := SubscribeExtension{
		router: r,
		dri:    dri,
	}
	switch cfg.Kind {
	case DefaultSubscription, BlockingSubscription:
		return &BlockingSubscriptionExtension{
			SubscribeExtension: ext,
		}, nil
	case ExternalSubscription:
		return &ExternalSubscriptionExtension{
			SubscribeExtension: ext,
		}, nil
	default:
		return nil, errors.New("this subscription kind is not implemented yet")
	}
}

func isSubscription(info *graphql.ResolveInfo) bool {
	return info != nil &&
		info.Path != nil &&
		info.Path.Prev == nil &&
		info.Operation != nil &&
		info.Operation.GetOperation() == ast.OperationTypeSubscription
}
