package grpc_test

import (
	"context"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientFieldResolve(t *testing.T) {
	data := []struct {
		title        string
		input        driver.FieldResolveInput
		grpcRequest  *proto.FieldResolveRequest
		grpcResponse *proto.FieldResolveResponse
		expected     driver.FieldResolveOutput
	}{
		{
			title: "MarshalingInput",
			input: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{
					FieldName: "field",
					Path: &types.ResponsePath{
						Key: "field",
						Prev: &types.ResponsePath{
							Key: "fieldPrev",
						},
					},
					ReturnType: &types.TypeRef{
						List: &types.TypeRef{
							NonNull: &types.TypeRef{
								Name: "String",
							},
						},
					},
					ParentType: &types.TypeRef{
						Name: "SomeType",
					},
					Operation: &types.OperationDefinition{
						Operation:           "query",
						Name:                "getFieldOfFieldPrev",
						VariableDefinitions: []types.VariableDefinition{},
						Directives: types.Directives{
							types.Directive{
								Name: "@somedir",
								Arguments: types.Arguments{
									"arg": "value",
									"astIntValue": &ast.IntValue{
										Value: "1",
									},
									"astFloatValue": &ast.FloatValue{
										Value: "1.0",
									},
									"astStringValue": &ast.StringValue{
										Value: "string",
									},
									"astBoolValue": &ast.BooleanValue{
										Value: true,
									},
									"astListValue": &ast.ListValue{
										Values: []ast.Value{
											&ast.IntValue{
												Value: "1",
											},
										},
									},
									"astObjectValue": &ast.ObjectValue{
										Fields: []*ast.ObjectField{
											{
												Name: &ast.Name{
													Value: "objectField",
												},
												Value: &ast.IntValue{
													Value: "1",
												},
											},
										},
									},
									"astVariable": &ast.Variable{
										Name: &ast.Name{
											Value: "someVar",
										},
									},
								},
							},
						},
						SelectionSet: types.Selections{
							types.Selection{
								Name: "subfield",
								Arguments: types.Arguments{
									"arg": "value",
								},
								Directives: types.Directives{
									types.Directive{
										Name: "@somedir",
										Arguments: types.Arguments{
											"arg": "value",
										},
									},
								},
								SelectionSet: types.Selections{
									types.Selection{
										Definition: &types.FragmentDefinition{
											Directives: types.Directives{
												types.Directive{
													Name: "@somedir",
													Arguments: types.Arguments{
														"arg": "value",
													},
												},
											},
											TypeCondition: types.TypeRef{
												Name: "SomeType",
											},
											SelectionSet: types.Selections{
												types.Selection{
													Name: "someField",
												},
											},
											VariableDefinitions: []types.VariableDefinition{
												types.VariableDefinition{
													Variable: types.Variable{
														Name: "variable",
													},
													DefaultValue: "default",
												},
											},
										},
									},
								},
							},
						},
					},
					VariableValues: map[string]interface{}{
						"var": "value",
					},
				},
			},
			grpcRequest: &proto.FieldResolveRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Info: &proto.FieldResolveInfo{
					FieldName: "field",
					Path: &proto.ResponsePath{
						Key: &proto.Value{
							TestValue: &proto.Value_S{
								S: "field",
							},
						},
						Prev: &proto.ResponsePath{
							Key: &proto.Value{
								TestValue: &proto.Value_S{
									S: "fieldPrev",
								},
							},
						},
					},
					ReturnType: &proto.TypeRef{
						TestTyperef: &proto.TypeRef_List{
							List: &proto.TypeRef{
								TestTyperef: &proto.TypeRef_NonNull{
									NonNull: &proto.TypeRef{
										TestTyperef: &proto.TypeRef_Name{
											Name: "String",
										},
									},
								},
							},
						},
					},
					ParentType: &proto.TypeRef{
						TestTyperef: &proto.TypeRef_Name{
							Name: "SomeType",
						},
					},
					Operation: &proto.OperationDefinition{
						Operation:           "query",
						Name:                "getFieldOfFieldPrev",
						VariableDefinitions: []*proto.VariableDefinition{},
						Directives: []*proto.Directive{
							&proto.Directive{
								Name: "@somedir",
								Arguments: map[string]*proto.Value{
									"arg": &proto.Value{
										TestValue: &proto.Value_S{
											S: "value",
										},
									},
									"astIntValue": &proto.Value{
										TestValue: &proto.Value_I{
											I: int64(1),
										},
									},
									"astFloatValue": &proto.Value{
										TestValue: &proto.Value_F{
											F: float64(1.0),
										},
									},
									"astStringValue": &proto.Value{
										TestValue: &proto.Value_S{
											S: "string",
										},
									},
									"astBoolValue": &proto.Value{
										TestValue: &proto.Value_B{
											B: true,
										},
									},
									"astListValue": &proto.Value{
										TestValue: &proto.Value_A{
											A: &proto.ArrayValue{
												Items: []*proto.Value{
													&proto.Value{
														TestValue: &proto.Value_I{
															I: int64(1),
														},
													},
												},
											},
										},
									},
									"astObjectValue": &proto.Value{
										TestValue: &proto.Value_O{
											O: &proto.ObjectValue{
												Props: map[string]*proto.Value{
													"objectField": &proto.Value{
														TestValue: &proto.Value_I{
															I: int64(1),
														},
													},
												},
											},
										},
									},
									"astVariable": &proto.Value{
										TestValue: &proto.Value_Variable{
											Variable: "someVar",
										},
									},
								},
							},
						},
						SelectionSet: []*proto.Selection{
							&proto.Selection{
								Name: "subfield",
								Arguments: map[string]*proto.Value{
									"arg": &proto.Value{
										TestValue: &proto.Value_S{
											S: "value",
										},
									},
								},
								Directives: []*proto.Directive{
									&proto.Directive{
										Name: "@somedir",
										Arguments: map[string]*proto.Value{
											"arg": &proto.Value{
												TestValue: &proto.Value_S{
													S: "value",
												},
											},
										},
									},
								},
								SelectionSet: []*proto.Selection{
									&proto.Selection{
										Definition: &proto.FragmentDefinition{
											Directives: []*proto.Directive{
												&proto.Directive{
													Name: "@somedir",
													Arguments: map[string]*proto.Value{
														"arg": &proto.Value{
															TestValue: &proto.Value_S{
																S: "value",
															},
														},
													},
												},
											},
											TypeCondition: &proto.TypeRef{
												TestTyperef: &proto.TypeRef_Name{
													Name: "SomeType",
												},
											},
											SelectionSet: []*proto.Selection{
												&proto.Selection{
													Name: "someField",
												},
											},
											VariableDefinitions: []*proto.VariableDefinition{
												&proto.VariableDefinition{
													Variable: &proto.Variable{
														Name: "variable",
													},
													DefaultValue: &proto.Value{
														TestValue: &proto.Value_S{
															S: "default",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					VariableValues: map[string]*proto.Value{
						"var": &proto.Value{
							TestValue: &proto.Value_S{
								S: "value",
							},
						},
					},
				},
				Protocol: new(proto.Value),
				Source:   new(proto.Value),
			},
			grpcResponse: &proto.FieldResolveResponse{
				Response: &proto.Value{
					TestValue: &proto.Value_S{
						S: "response",
					},
				},
			},
			expected: driver.FieldResolveOutput{
				Response: "response",
			},
		},
		{
			title: "MarshalingArbitrarySource",
			input: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{},
				Source: map[string]interface{}{
					"interfaceValue":   interface{}(1),
					"ptrValue":         func(i int) *int { return &i }(1),
					"intValue":         1,
					"int8Value":        int8(1),
					"int16Value":       int16(1),
					"int32Value":       int32(1),
					"int64Value":       int64(1),
					"uintValue":        uint(1),
					"uint8Value":       uint8(1),
					"uint16Value":      uint16(1),
					"uint32Value":      uint32(1),
					"uint64Value":      uint64(1),
					"float32Value":     float32(1.0),
					"float64Value":     float64(1.0),
					"stringValue":      "string",
					"boolValue":        true,
					"sliceValue":       []interface{}{1, "string"},
					"nilSlicePtrValue": (*[]interface{})(nil),
					"emptySlice":       []interface{}{},
					"arrayValue":       [2]interface{}{1, "string"},
					"bytesValue":       []byte("somebytes"),
					"structValue": struct {
						IntValue    int
						StringValue string
						TaggedValue string `json:"taggedValue"`
					}{1, "string", "tagged"},
					"mapValue": map[string]interface{}{
						"intValue":    1,
						"stringValue": "string",
					},
					"nilMapPtrValue":   (*map[string]interface{})(nil),
					"emptyMap":         map[string]interface{}{},
					"emptyIsMarshaled": nil,
				},
			},
			grpcRequest: &proto.FieldResolveRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Info:     &proto.FieldResolveInfo{},
				Protocol: new(proto.Value),
				Source: &proto.Value{
					TestValue: &proto.Value_O{
						O: &proto.ObjectValue{
							Props: map[string]*proto.Value{
								"interfaceValue": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"ptrValue": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"intValue": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"int8Value": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"int16Value": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"int32Value": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"int64Value": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"uintValue": &proto.Value{
									TestValue: &proto.Value_U{
										U: uint64(1),
									},
								},
								"uint8Value": &proto.Value{
									TestValue: &proto.Value_U{
										U: uint64(1),
									},
								},
								"uint16Value": &proto.Value{
									TestValue: &proto.Value_U{
										U: uint64(1),
									},
								},
								"uint32Value": &proto.Value{
									TestValue: &proto.Value_U{
										U: uint64(1),
									},
								},
								"uint64Value": &proto.Value{
									TestValue: &proto.Value_U{
										U: uint64(1),
									},
								},
								"float32Value": &proto.Value{
									TestValue: &proto.Value_F{
										F: float64(1.0),
									},
								},
								"float64Value": &proto.Value{
									TestValue: &proto.Value_F{
										F: float64(1.0),
									},
								},
								"stringValue": &proto.Value{
									TestValue: &proto.Value_S{
										S: "string",
									},
								},
								"boolValue": &proto.Value{
									TestValue: &proto.Value_B{
										B: true,
									},
								},
								"sliceValue": &proto.Value{
									TestValue: &proto.Value_A{
										A: &proto.ArrayValue{
											Items: []*proto.Value{
												&proto.Value{
													TestValue: &proto.Value_I{
														I: int64(1),
													},
												},
												&proto.Value{
													TestValue: &proto.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"nilSlicePtrValue": &proto.Value{
									TestValue: &proto.Value_A{},
								},
								"emptySlice": &proto.Value{
									TestValue: &proto.Value_A{
										A: &proto.ArrayValue{
											Items: make([]*proto.Value, 0),
										},
									},
								},
								"arrayValue": &proto.Value{
									TestValue: &proto.Value_A{
										A: &proto.ArrayValue{
											Items: []*proto.Value{
												&proto.Value{
													TestValue: &proto.Value_I{
														I: int64(1),
													},
												},
												&proto.Value{
													TestValue: &proto.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"bytesValue": &proto.Value{
									TestValue: &proto.Value_Any{
										Any: []byte("somebytes"),
									},
								},
								"structValue": &proto.Value{
									TestValue: &proto.Value_O{
										O: &proto.ObjectValue{
											Props: map[string]*proto.Value{
												"IntValue": &proto.Value{
													TestValue: &proto.Value_I{
														I: int64(1),
													},
												},
												"StringValue": &proto.Value{
													TestValue: &proto.Value_S{
														S: "string",
													},
												},
												"taggedValue": &proto.Value{
													TestValue: &proto.Value_S{
														S: "tagged",
													},
												},
											},
										},
									},
								},
								"mapValue": &proto.Value{
									TestValue: &proto.Value_O{
										O: &proto.ObjectValue{
											Props: map[string]*proto.Value{
												"intValue": &proto.Value{
													TestValue: &proto.Value_I{
														I: int64(1),
													},
												},
												"stringValue": &proto.Value{
													TestValue: &proto.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"nilMapPtrValue": &proto.Value{
									TestValue: &proto.Value_O{},
								},
								"emptyMap": &proto.Value{
									TestValue: &proto.Value_O{
										O: &proto.ObjectValue{
											Props: make(map[string]*proto.Value),
										},
									},
								},
								"emptyIsMarshaled": new(proto.Value),
							},
						},
					},
				},
			},
			grpcResponse: &proto.FieldResolveResponse{
				Response: &proto.Value{
					TestValue: &proto.Value_S{
						S: "response",
					},
				},
			},
			expected: driver.FieldResolveOutput{
				Response: "response",
			},
		},
		{
			title: "MarshalingObjectFieldsSource",
			input: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{},
				Source: []*ast.ObjectField{
					&ast.ObjectField{
						Name: &ast.Name{
							Value: "intField",
						},
						Value: &ast.IntValue{
							Value: "1",
						},
					},
					&ast.ObjectField{
						Name: &ast.Name{
							Value: "stringField",
						},
						Value: &ast.StringValue{
							Value: "string",
						},
					},
				},
			},
			grpcRequest: &proto.FieldResolveRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Info:     &proto.FieldResolveInfo{},
				Protocol: new(proto.Value),
				Source: &proto.Value{
					TestValue: &proto.Value_O{
						O: &proto.ObjectValue{
							Props: map[string]*proto.Value{
								"intField": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"stringField": &proto.Value{
									TestValue: &proto.Value_S{
										S: "string",
									},
								},
							},
						},
					},
				},
			},
			grpcResponse: &proto.FieldResolveResponse{
				Response: &proto.Value{
					TestValue: &proto.Value_S{
						S: "response",
					},
				},
			},
			expected: driver.FieldResolveOutput{
				Response: "response",
			},
		},
		{
			title: "UnmarshalingResponse",
			input: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{},
			},
			grpcRequest: &proto.FieldResolveRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Info:     &proto.FieldResolveInfo{},
				Protocol: new(proto.Value),
				Source:   new(proto.Value),
			},
			grpcResponse: &proto.FieldResolveResponse{
				Response: &proto.Value{
					TestValue: &proto.Value_O{
						O: &proto.ObjectValue{
							Props: map[string]*proto.Value{
								"intValue": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"uintValue": &proto.Value{
									TestValue: &proto.Value_U{
										U: uint64(1),
									},
								},
								"floatValue": &proto.Value{
									TestValue: &proto.Value_F{
										F: float64(1.0),
									},
								},
								"stringValue": &proto.Value{
									TestValue: &proto.Value_S{
										S: "string",
									},
								},
								"boolValue": &proto.Value{
									TestValue: &proto.Value_B{
										B: true,
									},
								},
								"sliceValue": &proto.Value{
									TestValue: &proto.Value_A{
										A: &proto.ArrayValue{
											Items: []*proto.Value{
												&proto.Value{
													TestValue: &proto.Value_I{
														I: int64(1),
													},
												},
												&proto.Value{
													TestValue: &proto.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"bytesValue": &proto.Value{
									TestValue: &proto.Value_Any{
										Any: []byte("somebytes"),
									},
								},
								"mapValue": &proto.Value{
									TestValue: &proto.Value_O{
										O: &proto.ObjectValue{
											Props: map[string]*proto.Value{
												"intValue": &proto.Value{
													TestValue: &proto.Value_I{
														I: int64(1),
													},
												},
												"stringValue": &proto.Value{
													TestValue: &proto.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"emptyIsMarshaled": new(proto.Value),
							},
						},
					},
				},
			},
			expected: driver.FieldResolveOutput{
				Response: map[string]interface{}{
					"intValue":    int64(1),
					"uintValue":   uint64(1),
					"floatValue":  float64(1.0),
					"stringValue": "string",
					"boolValue":   true,
					"sliceValue":  []interface{}{int64(1), "string"},
					"bytesValue":  []byte("somebytes"),
					"mapValue": map[string]interface{}{
						"intValue":    int64(1),
						"stringValue": "string",
					},
					"emptyIsMarshaled": nil,
				},
			},
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(
			tt.title,
			func(t *testing.T) {
				driverClientMock := new(driverClientMock)
				driverClientMock.On(
					"FieldResolve",
					mock.Anything,
					tt.grpcRequest,
				).Return(tt.grpcResponse, nil)
				client := grpc.Client{
					Client: driverClientMock,
				}
				out, err := client.FieldResolve(tt.input)
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, out)
				driverClientMock.AssertCalled(t, "FieldResolve", mock.Anything, tt.grpcRequest)
			},
		)
	}
}

func TestServerFieldResolve(t *testing.T) {
	data := []struct {
		title           string
		input           *proto.FieldResolveRequest
		handlerInput    driver.FieldResolveInput
		handlerResponse interface{}
		handlerError    error
		expected        *proto.FieldResolveResponse
		expectedErr     assert.ErrorAssertionFunc
	}{
		{
			title: "PassesCorrectSource",
			input: &proto.FieldResolveRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Info:     &proto.FieldResolveInfo{},
				Protocol: new(proto.Value),
				Source: &proto.Value{
					TestValue: &proto.Value_O{
						O: &proto.ObjectValue{
							Props: map[string]*proto.Value{
								"intValue": &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
								"uintValue": &proto.Value{
									TestValue: &proto.Value_U{
										U: uint64(1),
									},
								},
								"floatValue": &proto.Value{
									TestValue: &proto.Value_F{
										F: float64(1.0),
									},
								},
								"stringValue": &proto.Value{
									TestValue: &proto.Value_S{
										S: "string",
									},
								},
								"boolValue": &proto.Value{
									TestValue: &proto.Value_B{
										B: true,
									},
								},
								"sliceValue": &proto.Value{
									TestValue: &proto.Value_A{
										A: &proto.ArrayValue{
											Items: []*proto.Value{
												&proto.Value{
													TestValue: &proto.Value_I{
														I: int64(1),
													},
												},
												&proto.Value{
													TestValue: &proto.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"nilSliceValue": &proto.Value{
									TestValue: &proto.Value_A{},
								},
								"emptySlice": &proto.Value{
									TestValue: &proto.Value_A{
										A: &proto.ArrayValue{
											Items: make([]*proto.Value, 0),
										},
									},
								},
								"bytesValue": &proto.Value{
									TestValue: &proto.Value_Any{
										Any: []byte("somebytes"),
									},
								},
								"mapValue": &proto.Value{
									TestValue: &proto.Value_O{
										O: &proto.ObjectValue{
											Props: map[string]*proto.Value{
												"intValue": &proto.Value{
													TestValue: &proto.Value_I{
														I: int64(1),
													},
												},
												"stringValue": &proto.Value{
													TestValue: &proto.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"nilMapValue": &proto.Value{
									TestValue: &proto.Value_O{},
								},
								"emptyMap": &proto.Value{
									TestValue: &proto.Value_O{
										O: &proto.ObjectValue{
											Props: make(map[string]*proto.Value),
										},
									},
								},
								"emptyIsMarshaled": new(proto.Value),
							},
						},
					},
				},
			},
			handlerInput: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{},
				Source: map[string]interface{}{
					"intValue":      int64(1),
					"uintValue":     uint64(1),
					"floatValue":    float64(1.0),
					"stringValue":   "string",
					"boolValue":     true,
					"sliceValue":    []interface{}{int64(1), "string"},
					"nilSliceValue": ([]interface{})(nil),
					"emptySlice":    []interface{}{},
					"bytesValue":    []byte("somebytes"),
					"mapValue": map[string]interface{}{
						"intValue":    int64(1),
						"stringValue": "string",
					},
					"nilMapValue":      (map[string]interface{})(nil),
					"emptyMap":         map[string]interface{}{},
					"emptyIsMarshaled": nil,
				},
			},
			handlerResponse: "response",
			expected: &proto.FieldResolveResponse{
				Response: &proto.Value{
					TestValue: &proto.Value_S{
						S: "response",
					},
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title: "PassesCorrectInfoObject",
			input: &proto.FieldResolveRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Info: &proto.FieldResolveInfo{
					FieldName: "field",
					Path: &proto.ResponsePath{
						Key: &proto.Value{
							TestValue: &proto.Value_S{
								S: "field",
							},
						},
						Prev: &proto.ResponsePath{
							Key: &proto.Value{
								TestValue: &proto.Value_S{
									S: "fieldPrev",
								},
							},
						},
					},
					ReturnType: &proto.TypeRef{
						TestTyperef: &proto.TypeRef_List{
							List: &proto.TypeRef{
								TestTyperef: &proto.TypeRef_NonNull{
									NonNull: &proto.TypeRef{
										TestTyperef: &proto.TypeRef_Name{
											Name: "String",
										},
									},
								},
							},
						},
					},
					ParentType: &proto.TypeRef{
						TestTyperef: &proto.TypeRef_Name{
							Name: "SomeType",
						},
					},
					Operation: &proto.OperationDefinition{
						Operation:           "query",
						Name:                "getFieldOfFieldPrev",
						VariableDefinitions: []*proto.VariableDefinition{},
						Directives: []*proto.Directive{
							&proto.Directive{
								Name: "@somedir",
								Arguments: map[string]*proto.Value{
									"arg": &proto.Value{
										TestValue: &proto.Value_S{
											S: "value",
										},
									},
								},
							},
						},
						SelectionSet: []*proto.Selection{
							&proto.Selection{
								Name: "subfield",
								Arguments: map[string]*proto.Value{
									"arg": &proto.Value{
										TestValue: &proto.Value_S{
											S: "value",
										},
									},
								},
								Directives: []*proto.Directive{
									&proto.Directive{
										Name: "@somedir",
										Arguments: map[string]*proto.Value{
											"arg": &proto.Value{
												TestValue: &proto.Value_S{
													S: "value",
												},
											},
										},
									},
								},
								SelectionSet: []*proto.Selection{
									&proto.Selection{
										Definition: &proto.FragmentDefinition{
											Directives: []*proto.Directive{
												&proto.Directive{
													Name: "@somedir",
													Arguments: map[string]*proto.Value{
														"arg": &proto.Value{
															TestValue: &proto.Value_S{
																S: "value",
															},
														},
													},
												},
											},
											TypeCondition: &proto.TypeRef{
												TestTyperef: &proto.TypeRef_Name{
													Name: "SomeType",
												},
											},
											SelectionSet: []*proto.Selection{
												&proto.Selection{
													Name: "someField",
												},
											},
											VariableDefinitions: []*proto.VariableDefinition{
												&proto.VariableDefinition{
													Variable: &proto.Variable{
														Name: "variable",
													},
													DefaultValue: &proto.Value{
														TestValue: &proto.Value_S{
															S: "default",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					VariableValues: map[string]*proto.Value{
						"var": &proto.Value{
							TestValue: &proto.Value_S{
								S: "value",
							},
						},
					},
				},
				Protocol: new(proto.Value),
				Source:   new(proto.Value),
			},
			handlerInput: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{
					FieldName: "field",
					Path: &types.ResponsePath{
						Key: "field",
						Prev: &types.ResponsePath{
							Key: "fieldPrev",
						},
					},
					ReturnType: &types.TypeRef{
						List: &types.TypeRef{
							NonNull: &types.TypeRef{
								Name: "String",
							},
						},
					},
					ParentType: &types.TypeRef{
						Name: "SomeType",
					},
					Operation: &types.OperationDefinition{
						Operation:           "query",
						Name:                "getFieldOfFieldPrev",
						VariableDefinitions: []types.VariableDefinition{},
						Directives: types.Directives{
							types.Directive{
								Name: "@somedir",
								Arguments: types.Arguments{
									"arg": "value",
								},
							},
						},
						SelectionSet: types.Selections{
							types.Selection{
								Name: "subfield",
								Arguments: types.Arguments{
									"arg": "value",
								},
								Directives: types.Directives{
									types.Directive{
										Name: "@somedir",
										Arguments: types.Arguments{
											"arg": "value",
										},
									},
								},
								SelectionSet: types.Selections{
									types.Selection{
										Definition: &types.FragmentDefinition{
											Directives: types.Directives{
												types.Directive{
													Name: "@somedir",
													Arguments: types.Arguments{
														"arg": "value",
													},
												},
											},
											TypeCondition: types.TypeRef{
												Name: "SomeType",
											},
											SelectionSet: types.Selections{
												types.Selection{
													Name: "someField",
												},
											},
											VariableDefinitions: []types.VariableDefinition{
												types.VariableDefinition{
													Variable: types.Variable{
														Name: "variable",
													},
													DefaultValue: "default",
												},
											},
										},
									},
								},
							},
						},
					},
					VariableValues: map[string]interface{}{
						"var": "value",
					},
				},
			},
			handlerResponse: "response",
			expected: &proto.FieldResolveResponse{
				Response: &proto.Value{
					TestValue: &proto.Value_S{
						S: "response",
					},
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title: "HandlesIndexResponse",
			input: &proto.FieldResolveRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Info: &proto.FieldResolveInfo{
					FieldName: "field",
					Path: &proto.ResponsePath{
						Key: &proto.Value{
							TestValue: &proto.Value_S{
								S: "field",
							},
						},
						Prev: &proto.ResponsePath{
							Key: &proto.Value{
								TestValue: &proto.Value_I{
									I: int64(1),
								},
							},
						},
					},
				},
				Protocol: new(proto.Value),
				Source:   new(proto.Value),
			},
			handlerInput: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{
					FieldName: "field",
					Path: &types.ResponsePath{
						Key: "field",
						Prev: &types.ResponsePath{
							Key: int64(1),
						},
					},
				},
			},
			handlerResponse: "response",
			expected: &proto.FieldResolveResponse{
				Response: &proto.Value{
					TestValue: &proto.Value_S{
						S: "response",
					},
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title: "ServerReturnsVariableValue",
			input: &proto.FieldResolveRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Info: &proto.FieldResolveInfo{
					FieldName: "field",
					Operation: &proto.OperationDefinition{
						VariableDefinitions: []*proto.VariableDefinition{
							&proto.VariableDefinition{
								Variable: &proto.Variable{
									Name: "someVar",
								},
								DefaultValue: &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
							},
							&proto.VariableDefinition{
								Variable: &proto.Variable{
									Name: "someOtherVar",
								},
								DefaultValue: &proto.Value{
									TestValue: &proto.Value_I{
										I: int64(1),
									},
								},
							},
						},
						Directives: []*proto.Directive{
							&proto.Directive{
								Arguments: map[string]*proto.Value{
									"arg": &proto.Value{
										TestValue: &proto.Value_Variable{
											Variable: "someVar",
										},
									},
									"arg2": &proto.Value{
										TestValue: &proto.Value_Variable{
											Variable: "someOtherVar",
										},
									},
								},
							},
						},
					},
					VariableValues: map[string]*proto.Value{
						"someOtherVar": &proto.Value{
							TestValue: &proto.Value_I{
								I: int64(2),
							},
						},
					},
				},
				Protocol: new(proto.Value),
				Source:   new(proto.Value),
			},
			handlerInput: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{
					Operation: &types.OperationDefinition{
						VariableDefinitions: []types.VariableDefinition{
							types.VariableDefinition{
								Variable: types.Variable{
									Name: "someVar",
								},
								DefaultValue: int64(1),
							},
							types.VariableDefinition{
								Variable: types.Variable{
									Name: "someOtherVar",
								},
								DefaultValue: int64(1),
							},
						},
						Directives: types.Directives{
							types.Directive{
								Arguments: map[string]interface{}{
									"arg":  int64(1),
									"arg2": int64(2),
								},
							},
						},
					},
					VariableValues: map[string]interface{}{
						"someOtherVar": int64(2),
					},
					FieldName: "field",
				},
			},
			handlerResponse: "response",
			expected: &proto.FieldResolveResponse{
				Response: &proto.Value{
					TestValue: &proto.Value_S{
						S: "response",
					},
				},
			},
			expectedErr: assert.NoError,
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			fieldResolveMock := new(fieldResolveMock)
			fieldResolveMock.On("Handle", tt.handlerInput).Return(tt.handlerResponse, tt.handlerError)
			srv := grpc.Server{
				FieldResolveHandler: fieldResolveMock,
			}
			resp, err := srv.FieldResolve(context.Background(), tt.input)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, resp)
			fieldResolveMock.AssertCalled(t, "Handle", tt.handlerInput)
		})
	}
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		resp, err := srv.FieldResolve(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.NotEmpty(t, resp.Error.Msg)
	})
}
