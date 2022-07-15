package parser

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func defaultResolveFunc(types []*graphql.Object) func(graphql.ResolveTypeParams) *graphql.Object {
	// if data has defined __typename field we can use it in default implementation
	// to deduce an actual type of an object
	return func(p graphql.ResolveTypeParams) *graphql.Object {
		if resolvable, ok := p.Value.(TypeResolver); ok {
			return resolvable.ResolveType(p, types)
		}
		if len(types) == 0 {
			return nil
		}
		// If only one type available, return it
		if len(types) == 1 {
			return types[0]
		}
		v := reflect.ValueOf(p.Value)
		for (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsValid() {
			v = v.Elem()
		}
		if !v.IsValid() {
			return nil
		}
		switch v.Kind() {
		case reflect.Map:
			rtypename := v.MapIndex(reflect.ValueOf("__typename"))
			if rtypename.Kind() == reflect.Interface || rtypename.Kind() == reflect.Ptr {
				rtypename = rtypename.Elem()
			}
			if rtypename.IsValid() && rtypename.Kind() == reflect.String {
				typename := rtypename.String()
				for _, t := range types {
					if t.Name() == typename {
						return t
					}
				}
			}
		}
		return nil
	}
}

type TypeResolver interface {
	ResolveType(p graphql.ResolveTypeParams, types []*graphql.Object) *graphql.Object
}

func unionDefinition(p *Parser, u *ast.UnionDefinition) (t *graphql.Union, err error) {
	uCfg := graphql.UnionConfig{
		Name: u.Name.Value,
	}
	types := make([]*graphql.Object, 0, len(u.Types))
	setDescription(&uCfg.Description, u)
	t = graphql.NewUnion(uCfg)
	p.gqlTypeMap[t.Name()] = t
	for _, tt := range u.Types {
		gt, err := toGraphQLType(p, tt)
		if err != nil {
			return nil, err
		}
		// Invariant, according to spec
		// only object are allowed in union
		ot := gt.(*graphql.Object)
		types = append(types, ot)
	}
	if fn, ok := p.Unions[u.Name.Value]; ok {
		uCfg.ResolveType = fn
	} else {
		uCfg.ResolveType = defaultResolveFunc(types)
	}
	if len(types) >= 0 {
		uCfg.Types = types
	}
	*t = *graphql.NewUnion(uCfg)
	return t, nil
}
