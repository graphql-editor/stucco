package router

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
)

type routerContext int

// ContextKey returns shared router context from context associated with request
const ContextKey routerContext = 0

// Context context associated with request
type Context struct {
	Error error
}

type baseExtension struct{}

func (r baseExtension) Init(ctx context.Context, p *graphql.Params) context.Context {
	return ctx
}
func (r baseExtension) ParseDidStart(ctx context.Context) (context.Context, graphql.ParseFinishFunc) {
	return ctx, func(err error) {}
}

func (r baseExtension) ValidationDidStart(ctx context.Context) (context.Context, graphql.ValidationFinishFunc) {
	return ctx, func([]gqlerrors.FormattedError) {}
}

func (r baseExtension) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	return ctx, func(r *graphql.Result) {}
}

func (r baseExtension) ResolveFieldDidStart(ctx context.Context, info *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	return ctx, func(interface{}, error) {}
}

func (r baseExtension) HasResult(ctx context.Context) bool {
	return false
}

func (r baseExtension) GetResult(ctx context.Context) interface{} {
	return nil
}

type routerStartContext struct {
	baseExtension
}

func (r routerStartContext) Init(ctx context.Context, p *graphql.Params) context.Context {
	ctx = context.WithValue(ctx, ContextKey, &Context{})
	return ctx
}
func (r routerStartContext) Name() string { return "RouterStartExtension" }

type routerFinishContext struct {
	baseExtension
}

func (r routerFinishContext) Name() string { return "RouterFinishExtension" }

func (r routerFinishContext) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	assertRouterOk(ctx)
	return ctx, func(r *graphql.Result) {
		if err := getRouterError(ctx); err != nil {
			r.Data = nil
			r.Errors = nil
			panic(err)
		}
	}
}

func getRouterContext(ctx context.Context) *Context {
	if ctx == nil {
		return nil
	}
	return ctx.Value(ContextKey).(*Context)
}

func getRouterError(ctx context.Context) error {
	var err error
	if rtContext := getRouterContext(ctx); rtContext != nil && rtContext.Error != nil {
		err = rtContext.Error
	}
	return err
}

func assertRouterOk(ctx context.Context) {
	if err := getRouterError(ctx); err != nil {
		panic(err)
	}
}

func routerError(ctx context.Context, err error) {
	if rtContext := getRouterContext(ctx); rtContext.Error == nil {
		rtContext.Error = err
	}
}
