package router

import (
	"context"
	"errors"

	"github.com/graphql-go/graphql"
)

type authorizeExtension struct {
	baseExtension
	authorizeHandler func(*graphql.Params) (bool, error)
}

func (a authorizeExtension) Init(ctx context.Context, p *graphql.Params) context.Context {
	if a.authorizeHandler != nil {
		rtContext := getRouterContext(ctx)
		if rtContext.Error == nil {
			ok, err := a.authorizeHandler(p)
			if err != nil || !ok {
				if err == nil {
					err = errors.New("unauthorized")
				}
				rtContext.Error = err
			}
		}
	}
	return ctx
}
func (a authorizeExtension) Name() string { return "authorize extension" }
