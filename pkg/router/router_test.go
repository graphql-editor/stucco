package router_test

import (
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/drivertest"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRouter(t *testing.T) {
	defaultEnvironment := router.DefaultEnvironment()
	data := []struct {
		title       string
		in          router.Config
		expected    router.Router
		expectedErr assert.ErrorAssertionFunc
	}{
		{
			title: "Defaults",
			in: router.Config{
				Interfaces: map[string]router.InterfaceConfig{
					"SomeInterface": {ResolveType: types.Function{Name: "function"}},
				},
				Resolvers: map[string]router.ResolverConfig{
					"SomeType.field": {Resolve: types.Function{Name: "function"}},
				},
				Scalars: map[string]router.ScalarConfig{
					"SomeScalar": {
						Parse:     types.Function{Name: "function"},
						Serialize: types.Function{Name: "function"},
					},
				},
				Unions: map[string]router.UnionConfig{
					"SomeUnion": {ResolveType: types.Function{Name: "function"}},
				},
				Schema: `
interface SomeInterface{
	field: String
}
type SomeType {
	field: String
}
scalar SomeScalar
union SomeUnion = SomeType
schema {
	query: SomeType
}
`,
			},
			expected: router.Router{
				Interfaces: map[string]router.InterfaceConfig{
					"SomeInterface": {
						Environment: &defaultEnvironment,
						ResolveType: types.Function{Name: "function"},
					},
				},
				Resolvers: map[string]router.ResolverConfig{
					"SomeType.field": {
						Environment: &defaultEnvironment,
						Resolve:     types.Function{Name: "function"},
					},
				},
				Scalars: map[string]router.ScalarConfig{
					"SomeScalar": {
						Environment: &defaultEnvironment,
						Parse:       types.Function{Name: "function"},
						Serialize:   types.Function{Name: "function"},
					},
				},
				Unions: map[string]router.UnionConfig{
					"SomeUnion": {
						Environment: &defaultEnvironment,
						ResolveType: types.Function{Name: "function"},
					},
				},
			},
			expectedErr: assert.NoError,
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			mockDefaultDriver := driver.Config{
				Provider: defaultEnvironment.Provider,
				Runtime:  defaultEnvironment.Runtime,
			}
			mockDriver := new(drivertest.MockDriver)
			mockDriver.On("SetSecrets", mock.Anything).Return(driver.SetSecretsOutput{}, nil)
			driver.Register(mockDefaultDriver, mockDriver)
			out, err := router.NewRouter(tt.in)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected.Interfaces, out.Interfaces)
			assert.Equal(t, tt.expected.Resolvers, out.Resolvers)
			assert.Equal(t, tt.expected.Scalars, out.Scalars)
			assert.Equal(t, tt.expected.Unions, out.Unions)
			assert.NotNil(t, out.Schema.QueryType())
		})
	}
}
