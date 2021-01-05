package router

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/parser"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

type protocolKey int

// ProtocolKey used to pass extra data from protocol used for comunication
const ProtocolKey protocolKey = 0

type rawSubscriptionKey int

// RawSubscriptionKey used to pass instruction about how a subscription should be handled.
const RawSubscriptionKey rawSubscriptionKey = 0

// TypeMap contains defined GraphQL types
type TypeMap interface {
	Type(name string) graphql.Type
}

// Dispatch executes a resolution through a driver
type Dispatch struct {
	driver.Driver
	TypeMap TypeMap
}

func assertTypeRef(t *types.TypeRef) types.TypeRef {
	if t == nil {
		panic("assertion failed, TypeRef cannot be null here")
	}
	return *t
}

func makeTypeRefFromNamed(named *ast.Named) *types.TypeRef {
	if named == nil {
		return nil
	}
	return &types.TypeRef{Name: named.Name.Value}
}

func mustMakeTypeRefFromNamed(named *ast.Named) types.TypeRef {
	return assertTypeRef(makeTypeRefFromNamed(named))
}

func makeTypeRefFromType(t graphql.Type) *types.TypeRef {
	if t == nil {
		return nil
	}
	switch tt := t.(type) {
	case *graphql.NonNull:
		return &types.TypeRef{
			NonNull: makeTypeRefFromType(tt.OfType),
		}
	case *graphql.List:
		return &types.TypeRef{
			List: makeTypeRefFromType(tt.OfType),
		}
	}
	return &types.TypeRef{Name: t.Name()}
}

func newResposnePath(p *graphql.ResponsePath) *types.ResponsePath {
	if p == nil || p.Key == nil {
		return nil
	}
	return &types.ResponsePath{
		Prev: newResposnePath(p.Prev),
		Key:  p.Key,
	}
}

func makeArguments(args []*ast.Argument) types.Arguments {
	if len(args) == 0 {
		return nil
	}
	o := make(types.Arguments)
	for _, arg := range args {
		o[arg.Name.Value] = arg.Value
	}
	return o
}

func makeVariableDefinition(v *ast.VariableDefinition) types.VariableDefinition {
	if v == nil {
		return types.VariableDefinition{}
	}
	return types.VariableDefinition{
		Variable: types.Variable{
			Name: v.Variable.Name.Value,
		},
		DefaultValue: v.DefaultValue,
	}
}

func makeVariableDefintions(v []*ast.VariableDefinition) []types.VariableDefinition {
	if len(v) == 0 {
		return nil
	}
	r := make([]types.VariableDefinition, 0, len(v))
	for _, vv := range v {
		r = append(r, makeVariableDefinition(vv))
	}
	return r
}

func makeDirective(dir *ast.Directive) (d types.Directive) {
	if dir == nil {
		return
	}
	d.Name = dir.Name.Value
	d.Arguments = makeArguments(dir.Arguments)
	return
}

func makeDirectives(dirs []*ast.Directive) types.Directives {
	if len(dirs) == 0 {
		return nil
	}
	o := make(types.Directives, 0, len(dirs))
	for _, dir := range dirs {
		o = append(o, makeDirective(dir))
	}
	return o
}

func makeSelection(sel ast.Selection, fragments map[string]ast.Definition) (s types.Selection) {
	switch st := sel.(type) {
	case *ast.Field:
		s = types.Selection{
			Name:         st.Name.Value,
			Arguments:    makeArguments(st.Arguments),
			Directives:   makeDirectives(st.Directives),
			SelectionSet: makeSelections(st.SelectionSet, fragments),
		}
	case *ast.FragmentSpread:
		fdef := fragments[st.Name.Value].(*ast.FragmentDefinition)
		s = types.Selection{
			Directives: makeDirectives(st.Directives),
			Definition: &types.FragmentDefinition{
				TypeCondition:       mustMakeTypeRefFromNamed(fdef.TypeCondition),
				SelectionSet:        makeSelections(fdef.SelectionSet, fragments),
				Directives:          makeDirectives(fdef.Directives),
				VariableDefinitions: makeVariableDefintions(fdef.VariableDefinitions),
			},
		}
	case *ast.InlineFragment:
		s = types.Selection{
			Directives: makeDirectives(st.Directives),
			Definition: &types.FragmentDefinition{
				TypeCondition: mustMakeTypeRefFromNamed(st.TypeCondition),
				SelectionSet:  makeSelections(st.SelectionSet, fragments),
			},
		}
	}
	return
}

func makeSelections(selectionSet *ast.SelectionSet, fragments map[string]ast.Definition) types.Selections {
	if selectionSet == nil {
		return nil
	}
	selections := make(types.Selections, 0, len(selectionSet.Selections))
	for _, sel := range selectionSet.Selections {
		selections = append(selections, makeSelection(sel, fragments))
	}
	return selections
}

func makeOperationDefinition(odef *ast.OperationDefinition, fragments map[string]ast.Definition) *types.OperationDefinition {
	if odef == nil {
		return nil
	}
	name := ""
	if odef.Name != nil {
		name = odef.Name.Value
	}
	return &types.OperationDefinition{
		Operation:           odef.Operation,
		Name:                name,
		Directives:          makeDirectives(odef.Directives),
		VariableDefinitions: makeVariableDefintions(odef.VariableDefinitions),
		SelectionSet:        makeSelections(odef.SelectionSet, fragments),
	}
}

func buildFieldInfoParams(params graphql.ResolveInfo) driver.FieldResolveInfo {
	info := driver.FieldResolveInfo{
		FieldName:      params.FieldName,
		ReturnType:     makeTypeRefFromType(params.ReturnType),
		ParentType:     makeTypeRefFromType(params.ParentType),
		VariableValues: params.VariableValues,
		Path:           newResposnePath(params.Path),
		RootValue:      params.RootValue,
	}
	odef, ok := params.Operation.(*ast.OperationDefinition)
	if ok {
		info.Operation = makeOperationDefinition(odef, params.Fragments)
	}
	return info
}

// FieldResolve creates a function that calls implementation of field resolution through driver
func (d Dispatch) FieldResolve(rs ResolverConfig) func(params graphql.ResolveParams) (interface{}, error) {
	return func(params graphql.ResolveParams) (interface{}, error) {
		// short circuit subscription call
		if params.Context != nil {
			subCtx, ok := params.Context.Value(subscriptionExtensionKey).(*SubscribeContext)
			if ok && subCtx.IsSubscription {
				return nil, nil
			}
		}
		info := buildFieldInfoParams(params.Info)
		input := driver.FieldResolveInput{
			Function:  rs.Resolve,
			Source:    params.Source,
			Arguments: types.Arguments(params.Args),
			Info:      info,
		}
		if params.Context != nil {
			input.Protocol = params.Context.Value(ProtocolKey)
		}
		var err error
		out := d.Driver.FieldResolve(input)
		var i interface{}
		if err == nil {
			if out.Error != nil {
				err = fmt.Errorf(out.Error.Message)
			} else {
				i = out.Response
			}
		}
		if err != nil {
			err = errors.Wrap(err, rs.Resolve.Name)
		}
		return i, err
	}
}

func buildInterfaceInfoParams(params graphql.ResolveInfo) driver.InterfaceResolveTypeInfo {
	info := driver.InterfaceResolveTypeInfo{
		FieldName:      params.FieldName,
		ReturnType:     makeTypeRefFromType(params.ReturnType),
		ParentType:     makeTypeRefFromType(params.ParentType),
		VariableValues: params.VariableValues,
	}
	path := newResposnePath(params.Path)
	info.Path = path
	odef, ok := params.Operation.(*ast.OperationDefinition)
	if ok {
		info.Operation = makeOperationDefinition(odef, params.Fragments)
	}
	return info
}

// InterfaceResolveType creates a function that calls implementation of interface type resolution
func (d Dispatch) InterfaceResolveType(i InterfaceConfig) func(params graphql.ResolveTypeParams) *graphql.Object {
	return func(params graphql.ResolveTypeParams) *graphql.Object {
		input := driver.InterfaceResolveTypeInput{
			Function: i.ResolveType,
			Value:    params.Value,
			Info:     buildInterfaceInfoParams(params.Info),
		}
		var err error
		out := d.Driver.InterfaceResolveType(input)
		if out.Error != nil {
			err = fmt.Errorf(out.Error.Message)
		}
		var t *graphql.Object
		if err == nil {
			var ok bool
			t, ok = d.TypeMap.Type(out.Type.Name).(*graphql.Object)
			if !ok || t == nil {
				err = fmt.Errorf("\"out.Type.Name\" is not a valid type name")
			}
		}
		if err != nil {
			err = errors.Wrap(err, i.ResolveType.Name)
			panic(err)
		}
		return t
	}
}

// ScalarFunctions creates parse and serialize scalar functions that call implementation of scalar and parse through driver
func (d Dispatch) ScalarFunctions(s ScalarConfig) parser.ScalarFunctions {
	return parser.ScalarFunctions{
		Parse: func(v interface{}) interface{} {
			var err error
			out := d.Driver.ScalarParse(driver.ScalarParseInput{
				Function: s.Parse,
				Value:    v,
			})
			if err == nil {
				if out.Error != nil {
					err = fmt.Errorf(out.Error.Message)
				}
			}
			if err != nil {
				err = errors.Wrap(err, s.Parse.Name)
				// panic on error as there is no other way to
				// pass error from parse function to graphql-go
				panic(err)
			}
			return out.Response
		},
		Serialize: func(v interface{}) interface{} {
			var err error
			out := d.Driver.ScalarSerialize(driver.ScalarSerializeInput{
				Function: s.Serialize,
				Value:    v,
			})
			if err == nil {
				if out.Error != nil {
					err = fmt.Errorf(out.Error.Message)
				}
			}
			if err != nil {
				err = errors.Wrap(err, s.Parse.Name)
				// panic on error as there is no other way to
				// pass error from parse function to graphql-go
				panic(err)
			}
			return out.Response
		},
	}
}

func buildUnionInfoParams(params graphql.ResolveInfo) driver.UnionResolveTypeInfo {
	info := driver.UnionResolveTypeInfo{
		FieldName:      params.FieldName,
		ReturnType:     makeTypeRefFromType(params.ReturnType),
		ParentType:     makeTypeRefFromType(params.ParentType),
		VariableValues: params.VariableValues,
	}
	path := newResposnePath(params.Path)
	info.Path = path
	odef, ok := params.Operation.(*ast.OperationDefinition)
	if ok {
		info.Operation = makeOperationDefinition(odef, params.Fragments)
	}
	return info
}

// UnionResolveType creates a function that calls union resolution using driver
func (d Dispatch) UnionResolveType(u UnionConfig) func(params graphql.ResolveTypeParams) *graphql.Object {
	return func(params graphql.ResolveTypeParams) *graphql.Object {
		input := driver.UnionResolveTypeInput{
			Function: u.ResolveType,
			Value:    params.Value,
			Info:     buildUnionInfoParams(params.Info),
		}
		var err error
		out := d.Driver.UnionResolveType(input)
		if err == nil && out.Error != nil {
			err = fmt.Errorf(out.Error.Message)
		}
		var t *graphql.Object
		if err == nil {
			var ok bool
			t, ok = d.TypeMap.Type(out.Type.Name).(*graphql.Object)
			if !ok || t == nil {
				err = fmt.Errorf("\"out.Type.Name\" is not a valid type name")
			}
		}
		if err != nil {
			err = errors.Wrap(err, u.ResolveType.Name)
			panic(err)
		}
		return t
	}
}
