package parser

import (
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

func TestInterfaceDefinition(t *testing.T) {
	p := &Parser{
		gqlTypeMap: graphql.TypeMap{
			"FooConcrete": &graphql.Object{},
		},
		definitions: map[string]ast.Definition{
			"FooConcrete": &ast.ObjectDefinition{
				Name: &ast.Name{
					Value: "FooConcrete",
				},
				Interfaces: []*ast.Named{
					&ast.Named{
						Name: &ast.Name{
							Value: "Foo",
						},
					},
				},
			},
		},
	}
	data := []struct {
		in                   *ast.InterfaceDefinition
		out                  *graphql.Interface
		resolveTypeAssertion assert.ValueAssertionFunc
	}{
		{
			in: &ast.InterfaceDefinition{
				Name: &ast.Name{
					Value: "Foo",
				},
				Fields: []*ast.FieldDefinition{
					&ast.FieldDefinition{
						Name: &ast.Name{
							Value: "foo",
						},
						Type: &ast.Named{
							Name: &ast.Name{
								Value: "String",
							},
						},
					},
				},
			},
			out: graphql.NewInterface(graphql.InterfaceConfig{
				Name: "Foo",
				Fields: graphql.Fields{
					"foo": &graphql.Field{
						Type: graphql.String,
					},
				},
			}),
			resolveTypeAssertion: func(t assert.TestingT, i interface{}, rest ...interface{}) bool {
				return assert.IsType(t, i, graphql.ResolveTypeFn(func(graphql.ResolveTypeParams) *graphql.Object { return nil })) &&
					assert.Equal(t, p.gqlTypeMap["FooConcrete"], (i.(graphql.ResolveTypeFn))(graphql.ResolveTypeParams{}))
			},
		},
	}
	for _, tt := range data {
		i, err := interfaceDefinition(p, tt.in)
		assert.NoError(t, err)
		assert.Equal(t, tt.out.Name(), i.Name())
		assert.Len(t, i.Fields(), len(tt.out.Fields()))
		for k, v := range tt.out.Fields() {
			assert.Equal(t, v, i.Fields()[k])
		}
		tt.resolveTypeAssertion(t, i.ResolveType)
	}
}
