package router

import (
	"errors"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/parser"
	"github.com/graphql-editor/stucco/pkg/types"
)

type protocolKey int

// ProtocolKey used to pass extra data from protocol used for comunication
const ProtocolKey protocolKey = 0

type driDispatch struct {
	driver.Driver
	r *Router
}

func flattenArgValue(av ast.Value) interface{} {
	switch astValue := av.(type) {
	case *ast.Variable:
		return astValue.Name.Value
	case *ast.IntValue, *ast.FloatValue, *ast.StringValue, *ast.BooleanValue, *ast.EnumValue:
		return av.GetValue()
	case *ast.ListValue:
		arr := make([]interface{}, len(astValue.Values))
		for i := 0; i < len(arr); i++ {
			arr[i] = flattenArgValue(astValue.Values[i])
		}
		return arr
	case *ast.ObjectValue:
		obj := make(map[string]interface{})
		for _, f := range astValue.Fields {
			obj[f.Name.Value] = flattenArgValue(f.Value)
		}
		return obj
	}
	return av.GetValue()
}
func flattenArgObjectField(aof []*ast.ObjectField) interface{} {
	aa := make(map[string]interface{})
	for _, f := range aof {
		aa[f.Name.Value] = flattenArgValue(f.Value)
	}
	return aa
}

func (dri driDispatch) FieldResolve(rs ResolverConfig) func(params graphql.ResolveParams) (interface{}, error) {
	return func(params graphql.ResolveParams) (interface{}, error) {
		info, err := buildFieldInfoParams(params.Info)
		if err != nil {
			return nil, err
		}
		args := params.Args
		for k, v := range args {
			switch vt := v.(type) {
			case []*ast.ObjectField:
				args[k] = flattenArgObjectField(vt)
			case ast.Value:
				args[k] = flattenArgValue(vt)
			}
		}
		out, err := dri.Driver.FieldResolve(driver.FieldResolveInput{
			Function:  rs.Resolve,
			Source:    params.Source,
			Arguments: types.Arguments(args),
			Info:      info,
			Protocol:  params.Context.Value(ProtocolKey),
		})
		if err != nil || out.Error != nil {
			if err == nil {
				err = errors.New(out.Error.Message)
			}
			return nil, err
		}
		return out.Response, nil
	}
}

func (dri driDispatch) InterfaceResolveType(i InterfaceConfig) func(params graphql.ResolveTypeParams) *graphql.Object {
	return func(params graphql.ResolveTypeParams) *graphql.Object {
		iinfo, err := buildInterfaceInfoParams(params.Info)
		if err != nil {
			// Errors here panic so that graphql-go picks them up
			// in recovery, no other choice
			panic(err)
		}
		input := driver.InterfaceResolveTypeInput{
			Function: i.ResolveType,
			Value:    params.Value,
			Info:     iinfo,
		}
		out, err := dri.Driver.InterfaceResolveType(input)
		if err != nil || out.Error != nil || out.Type.Name == "" {
			if err == nil {
				if out.Error != nil {
					err = errors.New(out.Error.Message)
				} else {
					err = errors.New("missing type name in type resolution")
				}
			}
			panic(err)
		}
		t, ok := dri.r.Schema.Type(out.Type.Name).(*graphql.Object)
		if !ok {
			panic(fmt.Errorf("type %s is not an object", out.Type.Name))
		}
		return t
	}
}

func (dri driDispatch) ScalarFunctions(s ScalarConfig) parser.ScalarFunctions {
	return parser.ScalarFunctions{
		Parse: func(v interface{}) interface{} {
			out, err := dri.Driver.ScalarParse(driver.ScalarParseInput{
				Function: s.Parse,
				Value:    v,
			})
			if err == nil {
				if out.Error != nil {
					err = errors.New(out.Error.Message)
				}
			}
			if err != nil {
				// panic on error as there is no other way to
				// pass error from parse function to graphql-go
				panic(err)
			}
			return out.Response
		},
		Serialize: func(v interface{}) interface{} {
			out, err := dri.Driver.ScalarSerialize(driver.ScalarSerializeInput{
				Function: s.Serialize,
				Value:    v,
			})
			if err == nil {
				if out.Error != nil {
					err = errors.New(out.Error.Message)
				}
			}
			if err != nil {
				// panic on error as there is no other way to
				// pass error from parse function to graphql-go
				panic(err)
			}
			return out.Response
		},
	}
}

func (dri driDispatch) UnionResolveType(u UnionConfig) func(params graphql.ResolveTypeParams) *graphql.Object {
	return func(params graphql.ResolveTypeParams) *graphql.Object {
		uinfo, err := buildUnionInfoParams(params.Info)
		if err != nil {
			// Errors here panic so that graphql-go picks them up
			// in recovery, no other choice
			panic(err)
		}
		input := driver.UnionResolveTypeInput{
			Function: u.ResolveType,
			Value:    params.Value,
			Info:     uinfo,
		}
		out, err := dri.Driver.UnionResolveType(input)
		if err != nil || out.Error != nil || out.Type.Name == "" {
			if err == nil {
				if out.Error != nil {
					err = errors.New(out.Error.Message)
				} else {
					err = errors.New("missing type name in type resolution")
				}
			}
			panic(err)
		}
		t, ok := dri.r.Schema.Type(out.Type.Name).(*graphql.Object)
		if !ok {
			panic(fmt.Errorf("type %s is not an object", out.Type.Name))
		}
		return t
	}
}
