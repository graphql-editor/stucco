package parser

import (
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func defaultResolveFunc(types []*graphql.Object) func(graphql.ResolveTypeParams) *graphql.Object {
	return func(p graphql.ResolveTypeParams) *graphql.Object {
		if resolvable, ok := p.Value.(TypeResolver); ok {
			return resolvable.ResolveType(p, types)
		}
		if len(types) == 0 {
			return nil
		}
		// If only one type available, return in
		if len(types) == 1 {
			return types[0]
		}
		// Heavily experimental,
		// needs some serious testing.
		var v reflect.Value
		for v = reflect.ValueOf(p.Value); v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		}
		if !v.IsValid() {
			return nil
		}
		switch v.Kind() {
		case reflect.Struct:
			sn := v.Type().Name()
			for _, o := range types {
				// Match by type name first
				// if golang type name matches GraphQL type
				// name or {GraphQLTypeName}Type pattern return
				// that object
				capGqlName := strings.ToUpper(o.Name()[:1]) + string(o.Name()[1:])
				if capGqlName == sn || capGqlName+"Type" == sn {
					return o
				}
			}
			fallthrough
		case reflect.Map:
			// First check for __typename field presence and if matching type exists, return it
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
			// Try scoring and returning best match
			// TODO: for now it just does a very shallow check
			// without actually asserting that fields can
			// be assigned
			var o *graphql.Object
			score := -1
			for _, uo := range types {
				reqKeys := make([]string, 0, len(uo.Fields()))
				keys := make([]string, 0, len(uo.Fields()))
				for _, f := range uo.Fields() {
					if _, ok := f.Type.(*graphql.NonNull); ok {
						reqKeys = append(reqKeys, f.Name)
					} else {
						keys = append(keys, f.Name)
					}
				}
				newScore := matchScore(reqKeys, keys, v)
				if newScore > score {
					o = uo
					score = newScore
				}
			}
			return o
		}
		return nil
	}
}

type TypeResolver interface {
	ResolveType(p graphql.ResolveTypeParams, types []*graphql.Object) *graphql.Object
}

func matchScore(reqKeys []string, keys []string, v reflect.Value) int {
	score := len(keys) + len(reqKeys)
	matches := make(map[string]struct{}, 0)
	if v.Kind() == reflect.Map {
		for _, k := range v.MapKeys() {
			matches[k.Interface().(string)] = struct{}{}
		}
	} else {
		t := v.Type()
		getTag := func(f reflect.StructField) string {
			t := f.Tag.Get("graphql")
			if t == "" {
				t = f.Tag.Get("json")
			}
			if t != "" {
				t = strings.Split(t, ",")[0]
			}
			return t
		}
		// must be a struct
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			matches[f.Name] = struct{}{}
			matches[strings.ToLower(f.Name[:1])+f.Name[1:]] = struct{}{}
			if tag := getTag(f); tag != "" {
				matches[tag] = struct{}{}
			}
		}
	}
	for _, k := range reqKeys {
		if _, ok := matches[k]; ok {
			score++
		} else {
			// If required key is missing
			// this value does not match
			// the object at all
			return -1
		}
	}
	for _, k := range keys {
		if _, ok := matches[k]; ok {
			score++
		} else {
			score--
		}
	}
	return score
}

func unionDefinition(p *Parser, u *ast.UnionDefinition) (t *graphql.Union, err error) {
	uCfg := graphql.UnionConfig{
		Name:  u.Name.Value,
		Types: make([]*graphql.Object, 0, len(u.Types)),
	}
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
		uCfg.Types = append(uCfg.Types, ot)
	}
	if fn, ok := p.Unions[u.Name.Value]; ok {
		uCfg.ResolveType = fn
	} else {
		uCfg.ResolveType = defaultResolveFunc(uCfg.Types)
	}
	*t = *graphql.NewUnion(uCfg)
	return t, nil
}
