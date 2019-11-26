package parser

import (
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type namedDefinitionMock string

func (n namedDefinitionMock) GetOperation() string { return "" }

func (n namedDefinitionMock) GetVariableDefinitions() []*ast.VariableDefinition { return nil }

func (n namedDefinitionMock) GetSelectionSet() *ast.SelectionSet { return nil }

func (n namedDefinitionMock) GetKind() string { return "" }

func (n namedDefinitionMock) GetLoc() *ast.Location { return nil }

func (n namedDefinitionMock) GetName() *ast.Name { return &ast.Name{Value: string(n)} }

func (n namedDefinitionMock) Name() string { return string(n) }

func (n namedDefinitionMock) Description() string { return "" }

func (n namedDefinitionMock) String() string { return string(n) }

func (n namedDefinitionMock) Error() error { return nil }

func TestMakeField(t *testing.T) {
	p := &Parser{
		gqlTypeMap: graphql.TypeMap{
			"Foo": &graphql.Object{
				PrivateName: "Foo",
			},
			"Bar": &graphql.Object{
				PrivateName: "Bar",
			},
		},
		definitions: map[string]ast.Definition{
			"Foo": namedDefinitionMock("Foo"),
			"Bar": namedDefinitionMock("Bar"),
		},
		Config: Config{
			Resolvers: map[string]graphql.FieldResolveFn{
				"Bar.bar": func(p graphql.ResolveParams) (interface{}, error) {
					return "Bar.bar", nil
				},
			},
		},
	}
	data := []struct {
		in  *ast.FieldDefinition
		out *graphql.Field
	}{
		{
			in: &ast.FieldDefinition{
				Name: &ast.Name{
					Value: "foo",
				},
				Type: &ast.Named{
					Name: &ast.Name{
						Value: "Foo",
					},
				},
				Arguments: []*ast.InputValueDefinition{
					&ast.InputValueDefinition{
						Name: &ast.Name{
							Value: "input",
						},
						Type: &ast.Named{
							Name: &ast.Name{
								Value: "String",
							},
						},
					},
				},
			},
			out: &graphql.Field{
				Name: "foo",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Type: &graphql.Object{
					PrivateName: "Foo",
				},
			},
		},
		{
			in: &ast.FieldDefinition{
				Name: &ast.Name{
					Value: "bar",
				},
				Type: &ast.Named{
					Name: &ast.Name{
						Value: "Bar",
					},
				},
			},
			out: &graphql.Field{
				Name: "bar",
				Type: &graphql.Object{
					PrivateName: "Bar",
				},
			},
		},
	}
	for _, tt := range data {
		tn := tt.in.Type.(*ast.Named).Name.Value
		field, err := makeField(p, tn, tt.in)
		assert.NoError(t, err)
		assert.Equal(t, tt.out.Name, field.Name)
		assert.Equal(t, tt.out.Type, field.Type)
		if f, ok := p.Resolvers[utils.FieldName(tn, tt.in.Name.Value)]; ok {
			expectedOut, expectedErr := f(graphql.ResolveParams{})
			actualOut, actualErr := field.Resolve(graphql.ResolveParams{})
			assert.Equal(t, expectedOut, actualOut)
			assert.Equal(t, expectedErr, actualErr)
		}
	}
}

func TestMakeInputObjectField(t *testing.T) {
	p := &Parser{
		gqlTypeMap: graphql.TypeMap{
			"Foo": &graphql.InputObject{
				PrivateName: "Foo",
			},
		},
		definitions: map[string]ast.Definition{
			"Foo": namedDefinitionMock("Foo"),
		},
	}
	data := []struct {
		in  *ast.InputValueDefinition
		out *graphql.InputObjectFieldConfig
	}{
		{
			in: &ast.InputValueDefinition{
				Type: &ast.Named{
					Name: &ast.Name{
						Value: "Foo",
					},
				},
			},
			out: &graphql.InputObjectFieldConfig{
				Type: &graphql.InputObject{
					PrivateName: "Foo",
				},
			},
		},
	}
	for _, tt := range data {
		field, err := makeInputObjectField(p, tt.in)
		assert.NoError(t, err)
		assert.Equal(t, tt.out, field)
	}
}
