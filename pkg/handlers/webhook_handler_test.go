package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/graphql-editor/stucco/pkg/handlers"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateQuery(t *testing.T) {
	testType1 := graphql.NewObject(graphql.ObjectConfig{
		Name: "TestType1",
		Fields: graphql.Fields{
			"field": &graphql.Field{
				Type: graphql.String,
			},
		},
	})
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Types: []graphql.Type{
			testType1,
		},
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"field": &graphql.Field{
					Type: graphql.String,
				},
				"testType1Field1": &graphql.Field{
					Type: testType1,
				},
				"testType1Field2": &graphql.Field{
					Type: testType1,
					Args: graphql.FieldConfigArgument{
						"arg1": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
						"arg2": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
				},
			},
		}),
	})
	cfg := handlers.Config{
		Schema: &schema,
		RouterConfig: router.Config{
			Resolvers: map[string]router.ResolverConfig{
				"Query.testType1Field2": {
					Webhook: &router.WebhookConfig{
						Pattern: "/{arg1}/{arg2}",
					},
				},
			},
		},
	}
	require.NoError(t, err)
	q, err := handlers.CreateQuery(cfg, &http.Request{
		URL: &url.URL{
			Path: "/webhook/query/field",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "query{field}", q)

	q, err = handlers.CreateQuery(cfg, &http.Request{
		URL: &url.URL{
			Path: "/webhook/query/testType1Field1/field",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "query{testType1Field1{field}}", q)

	q, err = handlers.CreateQuery(cfg, &http.Request{
		URL: &url.URL{
			Path: "/webhook/query/testType1Field2/1/abc/field",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "query{testType1Field2(arg1: 1 arg2: \"abc\"){field}}", q)
}
