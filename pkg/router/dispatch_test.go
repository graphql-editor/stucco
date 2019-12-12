package router_test

import (
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/drivertest"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDispatchFieldResolve(t *testing.T) {
	data := []struct {
		title        string
		config       router.ResolverConfig
		params       graphql.ResolveParams
		driverInput  driver.FieldResolveInput
		driverOutput driver.FieldResolveOutput
		driverError  error
		expected     interface{}
		expectedErr  assert.ErrorAssertionFunc
	}{
		{
			title: "CallsFieldResolve",
			config: router.ResolverConfig{
				Resolve: types.Function{
					Name: "function",
				},
			},
			params: graphql.ResolveParams{
				Source: "source data",
				Args: map[string]interface{}{
					"arg": "value",
				},
				Info: graphql.ResolveInfo{
					FieldName: "field",
				},
			},
			driverInput: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Source: "source data",
				Arguments: map[string]interface{}{
					"arg": "value",
				},
				Info: driver.FieldResolveInfo{
					FieldName: "field",
				},
			},
			driverOutput: driver.FieldResolveOutput{
				Response: "response",
			},
			expected:    "response",
			expectedErr: assert.NoError,
		},
		{
			title: "ParsesInfoFields",
			config: router.ResolverConfig{
				Resolve: types.Function{
					Name: "function",
				},
			},
			params: graphql.ResolveParams{
				Source: "source data",
				Args: map[string]interface{}{
					"arg": "value",
				},
				Info: graphql.ResolveInfo{
					FieldName: "field",
					Fragments: map[string]ast.Definition{
						"SomeFragment": &ast.FragmentDefinition{
							TypeCondition: &ast.Named{
								Name: &ast.Name{
									Value: "SomeType",
								},
							},
							SelectionSet: &ast.SelectionSet{
								Selections: []ast.Selection{
									&ast.Field{
										Name: &ast.Name{
											Value: "someTypeField",
										},
									},
								},
							},
						},
					},
					Path: &graphql.ResponsePath{
						Key: "fieldPrev",
						Prev: &graphql.ResponsePath{
							Key: 1,
						},
					},
					ReturnType: &graphql.List{
						OfType: &graphql.NonNull{
							OfType: &graphql.Object{
								PrivateName: "SomeType",
							},
						},
					},
					ParentType: &graphql.Object{
						PrivateName: "SomeParentType",
					},
					Operation: &ast.OperationDefinition{
						Directives: []*ast.Directive{
							&ast.Directive{
								Name: &ast.Name{
									Value: "@someDir",
								},
								Arguments: []*ast.Argument{
									&ast.Argument{
										Name: &ast.Name{
											Value: "arg",
										},
										Value: &ast.Variable{
											Name: &ast.Name{
												Value: "var",
											},
										},
									},
								},
							},
						},
						Name: &ast.Name{
							Value: "operationName",
						},
						Operation: "query",
						SelectionSet: &ast.SelectionSet{
							Selections: []ast.Selection{
								&ast.Field{
									Name: &ast.Name{
										Value: "field",
									},
								},
								&ast.FragmentSpread{
									Name: &ast.Name{
										Value: "SomeFragment",
									},
								},
								&ast.InlineFragment{
									TypeCondition: &ast.Named{
										Name: &ast.Name{
											Value: "SomeType",
										},
									},
									SelectionSet: &ast.SelectionSet{
										Selections: []ast.Selection{
											&ast.Field{
												Name: &ast.Name{
													Value: "someTypeField",
												},
											},
										},
									},
								},
							},
						},
						VariableDefinitions: []*ast.VariableDefinition{
							&ast.VariableDefinition{
								Variable: &ast.Variable{
									Name: &ast.Name{
										Value: "someVar",
									},
								},
								DefaultValue: &ast.IntValue{
									Value: "1",
								},
							},
						},
					},
				},
			},
			driverInput: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Source: "source data",
				Arguments: map[string]interface{}{
					"arg": "value",
				},
				Info: driver.FieldResolveInfo{
					FieldName: "field",
					Path: &types.ResponsePath{
						Key: "fieldPrev",
						Prev: &types.ResponsePath{
							Key: 1,
						},
					},
					ReturnType: &types.TypeRef{
						List: &types.TypeRef{
							NonNull: &types.TypeRef{
								Name: "SomeType",
							},
						},
					},
					ParentType: &types.TypeRef{
						Name: "SomeParentType",
					},
					Operation: &types.OperationDefinition{
						Directives: types.Directives{
							types.Directive{
								Name: "@someDir",
								Arguments: types.Arguments{
									"arg": &ast.Variable{
										Name: &ast.Name{
											Value: "var",
										},
									},
								},
							},
						},
						Name:      "operationName",
						Operation: "query",
						SelectionSet: types.Selections{
							types.Selection{
								Name: "field",
							},
							types.Selection{
								Definition: &types.FragmentDefinition{
									TypeCondition: types.TypeRef{
										Name: "SomeType",
									},
									SelectionSet: types.Selections{
										types.Selection{
											Name: "someTypeField",
										},
									},
								},
							},
							types.Selection{
								Definition: &types.FragmentDefinition{
									TypeCondition: types.TypeRef{
										Name: "SomeType",
									},
									SelectionSet: types.Selections{
										types.Selection{
											Name: "someTypeField",
										},
									},
								},
							},
						},
						VariableDefinitions: []types.VariableDefinition{
							types.VariableDefinition{
								Variable: types.Variable{
									Name: "someVar",
								},
								DefaultValue: &ast.IntValue{
									Value: "1",
								},
							},
						},
					},
				},
			},
			driverOutput: driver.FieldResolveOutput{
				Response: "response",
			},
			expected:    "response",
			expectedErr: assert.NoError,
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			mockDriver := new(drivertest.MockDriver)
			mockDriver.On("FieldResolve", tt.driverInput).Return(tt.driverOutput, tt.driverError)
			out, err := router.Dispatch{Driver: mockDriver}.FieldResolve(tt.config)(tt.params)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
}

type typeMapMock struct {
	mock.Mock
}

func (m *typeMapMock) Type(name string) graphql.Type {
	return m.Called(name).Get(0).(graphql.Type)
}

func TestDispatchInterfaceResolveType(t *testing.T) {
	data := []struct {
		title         string
		config        router.InterfaceConfig
		params        graphql.ResolveTypeParams
		driverInput   driver.InterfaceResolveTypeInput
		driverOutput  driver.InterfaceResolveTypeOutput
		driverError   error
		typeMapInput  string
		schema        graphql.Schema
		expected      *graphql.Object
		expectedPanic func(assert.TestingT, assert.PanicTestFunc, ...interface{}) bool
	}{
		{
			title: "CallsInterfaceResolveType",
			config: router.InterfaceConfig{
				ResolveType: types.Function{
					Name: "function",
				},
			},
			params: graphql.ResolveTypeParams{
				Value: "Value",
				Info: graphql.ResolveInfo{
					FieldName: "field",
				},
			},
			driverInput: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "Value",
				Info: driver.InterfaceResolveTypeInfo{
					FieldName: "field",
				},
			},
			driverOutput: driver.InterfaceResolveTypeOutput{
				Type: types.TypeRef{
					Name: "SomeType",
				},
			},
			typeMapInput: "SomeType",
			expected: &graphql.Object{
				PrivateName: "SomeType",
			},
			expectedPanic: assert.NotPanics,
		},
		{
			title: "PanicsOnErrorOutput",
			config: router.InterfaceConfig{
				ResolveType: types.Function{
					Name: "function",
				},
			},
			params: graphql.ResolveTypeParams{
				Value: "Value",
				Info: graphql.ResolveInfo{
					FieldName: "field",
				},
			},
			driverInput: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "Value",
				Info: driver.InterfaceResolveTypeInfo{
					FieldName: "field",
				},
			},
			driverOutput: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "some error",
				},
			},
			expectedPanic: assert.Panics,
		},
		{
			title: "PanicsOnBadType",
			config: router.InterfaceConfig{
				ResolveType: types.Function{
					Name: "function",
				},
			},
			params: graphql.ResolveTypeParams{
				Value: "Value",
				Info: graphql.ResolveInfo{
					FieldName: "field",
				},
			},
			driverInput: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "Value",
				Info: driver.InterfaceResolveTypeInfo{
					FieldName: "field",
				},
			},
			driverOutput: driver.InterfaceResolveTypeOutput{
				Type: types.TypeRef{
					Name: "SomeType",
				},
			},
			typeMapInput:  "SomeType",
			expectedPanic: assert.Panics,
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			mockDriver := new(drivertest.MockDriver)
			mockDriver.On("InterfaceResolveType", tt.driverInput).Return(tt.driverOutput, tt.driverError)
			mockTypeMap := new(typeMapMock)
			mockTypeMap.On("Type", tt.typeMapInput).Return(tt.expected)
			var out *graphql.Object
			tt.expectedPanic(t, func() {
				out = router.Dispatch{
					Driver:  mockDriver,
					TypeMap: mockTypeMap,
				}.InterfaceResolveType(tt.config)(tt.params)
			})
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestDispatchUnionResolveType(t *testing.T) {
	data := []struct {
		title         string
		config        router.UnionConfig
		params        graphql.ResolveTypeParams
		driverInput   driver.UnionResolveTypeInput
		driverOutput  driver.UnionResolveTypeOutput
		driverError   error
		typeMapInput  string
		schema        graphql.Schema
		expected      *graphql.Object
		expectedPanic func(assert.TestingT, assert.PanicTestFunc, ...interface{}) bool
	}{
		{
			title: "CallsUnionResolveType",
			config: router.UnionConfig{
				ResolveType: types.Function{
					Name: "function",
				},
			},
			params: graphql.ResolveTypeParams{
				Value: "Value",
				Info: graphql.ResolveInfo{
					FieldName: "field",
				},
			},
			driverInput: driver.UnionResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "Value",
				Info: driver.UnionResolveTypeInfo{
					FieldName: "field",
				},
			},
			driverOutput: driver.UnionResolveTypeOutput{
				Type: types.TypeRef{
					Name: "SomeType",
				},
			},
			typeMapInput: "SomeType",
			expected: &graphql.Object{
				PrivateName: "SomeType",
			},
			expectedPanic: assert.NotPanics,
		},
		{
			title: "PanicsOnErrorOutput",
			config: router.UnionConfig{
				ResolveType: types.Function{
					Name: "function",
				},
			},
			params: graphql.ResolveTypeParams{
				Value: "Value",
				Info: graphql.ResolveInfo{
					FieldName: "field",
				},
			},
			driverInput: driver.UnionResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "Value",
				Info: driver.UnionResolveTypeInfo{
					FieldName: "field",
				},
			},
			driverOutput: driver.UnionResolveTypeOutput{
				Error: &driver.Error{
					Message: "some error",
				},
			},
			expectedPanic: assert.Panics,
		},
		{
			title: "PanicsOnBadType",
			config: router.UnionConfig{
				ResolveType: types.Function{
					Name: "function",
				},
			},
			params: graphql.ResolveTypeParams{
				Value: "Value",
				Info: graphql.ResolveInfo{
					FieldName: "field",
				},
			},
			driverInput: driver.UnionResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "Value",
				Info: driver.UnionResolveTypeInfo{
					FieldName: "field",
				},
			},
			driverOutput: driver.UnionResolveTypeOutput{
				Type: types.TypeRef{
					Name: "SomeType",
				},
			},
			typeMapInput:  "SomeType",
			expectedPanic: assert.Panics,
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			mockDriver := new(drivertest.MockDriver)
			mockDriver.On("UnionResolveType", tt.driverInput).Return(tt.driverOutput, tt.driverError)
			mockTypeMap := new(typeMapMock)
			mockTypeMap.On("Type", tt.typeMapInput).Return(tt.expected)
			var out *graphql.Object
			tt.expectedPanic(t, func() {
				out = router.Dispatch{
					Driver:  mockDriver,
					TypeMap: mockTypeMap,
				}.UnionResolveType(tt.config)(tt.params)
			})
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestDispatchScalarParse(t *testing.T) {
	data := []struct {
		title         string
		config        router.ScalarConfig
		input         interface{}
		driverInput   driver.ScalarParseInput
		driverOutput  driver.ScalarParseOutput
		expected      interface{}
		expectedPanic func(assert.TestingT, assert.PanicTestFunc, ...interface{}) bool
	}{
		{
			title: "CallsScalarParse",
			config: router.ScalarConfig{
				Parse: types.Function{
					Name: "function",
				},
			},
			input: "scalar val",
			driverInput: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "scalar val",
			},
			driverOutput: driver.ScalarParseOutput{
				Response: "scalar response",
			},
			expected:      "scalar response",
			expectedPanic: assert.NotPanics,
		},
		{
			title: "PanicsOnError",
			config: router.ScalarConfig{
				Parse: types.Function{
					Name: "function",
				},
			},
			input: "scalar val",
			driverInput: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "scalar val",
			},
			driverOutput: driver.ScalarParseOutput{
				Error: &driver.Error{
					Message: "some error",
				},
			},
			expectedPanic: assert.Panics,
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			mockDriver := new(drivertest.MockDriver)
			mockDriver.On("ScalarParse", tt.driverInput).Return(tt.driverOutput, nil)
			var out interface{}
			tt.expectedPanic(t, func() {
				out = router.Dispatch{
					Driver: mockDriver,
				}.ScalarFunctions(tt.config).Parse(tt.input)
			})
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestDispatchScalarSerialize(t *testing.T) {
	data := []struct {
		title         string
		config        router.ScalarConfig
		input         interface{}
		driverInput   driver.ScalarSerializeInput
		driverOutput  driver.ScalarSerializeOutput
		expected      interface{}
		expectedPanic func(assert.TestingT, assert.PanicTestFunc, ...interface{}) bool
	}{
		{
			title: "CallsScalarSerialize",
			config: router.ScalarConfig{
				Serialize: types.Function{
					Name: "function",
				},
			},
			input: "scalar val",
			driverInput: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "scalar val",
			},
			driverOutput: driver.ScalarSerializeOutput{
				Response: "scalar response",
			},
			expected:      "scalar response",
			expectedPanic: assert.NotPanics,
		},
		{
			title: "PanicsOnError",
			config: router.ScalarConfig{
				Serialize: types.Function{
					Name: "function",
				},
			},
			input: "scalar val",
			driverInput: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
				Value: "scalar val",
			},
			driverOutput: driver.ScalarSerializeOutput{
				Error: &driver.Error{
					Message: "some error",
				},
			},
			expectedPanic: assert.Panics,
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			mockDriver := new(drivertest.MockDriver)
			mockDriver.On("ScalarSerialize", tt.driverInput).Return(tt.driverOutput, nil)
			var out interface{}
			tt.expectedPanic(t, func() {
				out = router.Dispatch{
					Driver: mockDriver,
				}.ScalarFunctions(tt.config).Serialize(tt.input)
			})
			assert.Equal(t, tt.expected, out)
		})
	}
}
