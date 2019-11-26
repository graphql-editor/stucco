package parser

import (
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

func TestEnumDefinition(t *testing.T) {
	data := []struct {
		in  *ast.EnumDefinition
		out *graphql.Enum
	}{
		{
			in: &ast.EnumDefinition{
				Name: &ast.Name{
					Value: "Foo",
				},
				Values: []*ast.EnumValueDefinition{
					&ast.EnumValueDefinition{
						Name: &ast.Name{
							Value: "Fooo",
						},
					},
					&ast.EnumValueDefinition{
						Name: &ast.Name{
							Value: "Foooo",
						},
					},
					&ast.EnumValueDefinition{
						Name: &ast.Name{
							Value: "Fooooo",
						},
					},
				},
			},
			out: graphql.NewEnum(graphql.EnumConfig{
				Name: "Foo",
				Values: graphql.EnumValueConfigMap{
					"Fooo": &graphql.EnumValueConfig{
						Value: "Fooo",
					},
					"Foooo": &graphql.EnumValueConfig{
						Value: "Foooo",
					},
					"Fooooo": &graphql.EnumValueConfig{
						Value: "Fooooo",
					},
				},
			}),
		},
	}
	for _, tt := range data {
		p := Parser{
			gqlTypeMap: make(graphql.TypeMap),
		}
		e, err := enumDefinition(&p, tt.in)
		assert.NoError(t, err)
		assert.Equal(t, e.Name(), tt.out.Name())
		assert.ElementsMatch(t, tt.out.Values(), e.Values())
		assert.Contains(t, p.gqlTypeMap, "Foo")
	}
}
