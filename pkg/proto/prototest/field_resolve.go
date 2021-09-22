package prototest

import (
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
	"github.com/graphql-go/graphql/language/ast"
)

// FieldResolveClientTest is basic struct for testing clients implementing proto
type FieldResolveClientTest struct {
	Title         string
	Input         driver.FieldResolveInput
	ProtoRequest  *protoMessages.FieldResolveRequest
	ProtoResponse *protoMessages.FieldResolveResponse
	Expected      driver.FieldResolveOutput
}

// FieldResolveClientTestData is a data for testing field resolution of proto clients
func FieldResolveClientTestData() []FieldResolveClientTest {
	return []FieldResolveClientTest{
		{
			Title: "MarshalingInput",
			Input: driver.FieldResolveInput{
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
						Operation: "query",
						Name:      "getFieldOfFieldPrev",
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
												{
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
			ProtoRequest: &protoMessages.FieldResolveRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Info: &protoMessages.FieldResolveInfo{
					FieldName: "field",
					Path: &protoMessages.ResponsePath{
						Key: &protoMessages.Value{
							TestValue: &protoMessages.Value_S{
								S: "field",
							},
						},
						Prev: &protoMessages.ResponsePath{
							Key: &protoMessages.Value{
								TestValue: &protoMessages.Value_S{
									S: "fieldPrev",
								},
							},
						},
					},
					ReturnType: &protoMessages.TypeRef{
						TestTyperef: &protoMessages.TypeRef_List{
							List: &protoMessages.TypeRef{
								TestTyperef: &protoMessages.TypeRef_NonNull{
									NonNull: &protoMessages.TypeRef{
										TestTyperef: &protoMessages.TypeRef_Name{
											Name: "String",
										},
									},
								},
							},
						},
					},
					ParentType: &protoMessages.TypeRef{
						TestTyperef: &protoMessages.TypeRef_Name{
							Name: "SomeType",
						},
					},
					Operation: &protoMessages.OperationDefinition{
						Operation: "query",
						Name:      "getFieldOfFieldPrev",
						Directives: []*protoMessages.Directive{
							{
								Name: "@somedir",
								Arguments: map[string]*protoMessages.Value{
									"arg": {
										TestValue: &protoMessages.Value_S{
											S: "value",
										},
									},
									"astIntValue": {
										TestValue: &protoMessages.Value_I{
											I: int64(1),
										},
									},
									"astFloatValue": {
										TestValue: &protoMessages.Value_F{
											F: float64(1.0),
										},
									},
									"astStringValue": {
										TestValue: &protoMessages.Value_S{
											S: "string",
										},
									},
									"astBoolValue": {
										TestValue: &protoMessages.Value_B{
											B: true,
										},
									},
									"astListValue": {
										TestValue: &protoMessages.Value_A{
											A: &protoMessages.ArrayValue{
												Items: []*protoMessages.Value{
													{
														TestValue: &protoMessages.Value_I{
															I: int64(1),
														},
													},
												},
											},
										},
									},
									"astObjectValue": {
										TestValue: &protoMessages.Value_O{
											O: &protoMessages.ObjectValue{
												Props: map[string]*protoMessages.Value{
													"objectField": {
														TestValue: &protoMessages.Value_I{
															I: int64(1),
														},
													},
												},
											},
										},
									},
									"astVariable": {
										TestValue: &protoMessages.Value_Variable{
											Variable: "someVar",
										},
									},
								},
							},
						},
						SelectionSet: []*protoMessages.Selection{
							{
								Name: "subfield",
								Arguments: map[string]*protoMessages.Value{
									"arg": {
										TestValue: &protoMessages.Value_S{
											S: "value",
										},
									},
								},
								Directives: []*protoMessages.Directive{
									{
										Name: "@somedir",
										Arguments: map[string]*protoMessages.Value{
											"arg": {
												TestValue: &protoMessages.Value_S{
													S: "value",
												},
											},
										},
									},
								},
								SelectionSet: []*protoMessages.Selection{
									{
										Definition: &protoMessages.FragmentDefinition{
											Directives: []*protoMessages.Directive{
												{
													Name: "@somedir",
													Arguments: map[string]*protoMessages.Value{
														"arg": {
															TestValue: &protoMessages.Value_S{
																S: "value",
															},
														},
													},
												},
											},
											TypeCondition: &protoMessages.TypeRef{
												TestTyperef: &protoMessages.TypeRef_Name{
													Name: "SomeType",
												},
											},
											SelectionSet: []*protoMessages.Selection{
												{
													Name: "someField",
												},
											},
											VariableDefinitions: []*protoMessages.VariableDefinition{
												{
													Variable: &protoMessages.Variable{
														Name: "variable",
													},
													DefaultValue: &protoMessages.Value{
														TestValue: &protoMessages.Value_S{
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
					VariableValues: map[string]*protoMessages.Value{
						"var": {
							TestValue: &protoMessages.Value_S{
								S: "value",
							},
						},
					},
				},
				Source: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				Protocol: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				SubscriptionPayload: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
			},
			ProtoResponse: &protoMessages.FieldResolveResponse{
				Response: &protoMessages.Value{
					TestValue: &protoMessages.Value_S{
						S: "response",
					},
				},
			},
			Expected: driver.FieldResolveOutput{
				Response: "response",
			},
		},
		{
			Title: "MarshalingArbitrarySource",
			Input: driver.FieldResolveInput{
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
			ProtoRequest: &protoMessages.FieldResolveRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Info: &protoMessages.FieldResolveInfo{},
				Source: &protoMessages.Value{
					TestValue: &protoMessages.Value_O{
						O: &protoMessages.ObjectValue{
							Props: map[string]*protoMessages.Value{
								"interfaceValue": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"ptrValue": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"intValue": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"int8Value": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"int16Value": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"int32Value": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"int64Value": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"uintValue": {
									TestValue: &protoMessages.Value_U{
										U: uint64(1),
									},
								},
								"uint8Value": {
									TestValue: &protoMessages.Value_U{
										U: uint64(1),
									},
								},
								"uint16Value": {
									TestValue: &protoMessages.Value_U{
										U: uint64(1),
									},
								},
								"uint32Value": {
									TestValue: &protoMessages.Value_U{
										U: uint64(1),
									},
								},
								"uint64Value": {
									TestValue: &protoMessages.Value_U{
										U: uint64(1),
									},
								},
								"float32Value": {
									TestValue: &protoMessages.Value_F{
										F: float64(1.0),
									},
								},
								"float64Value": {
									TestValue: &protoMessages.Value_F{
										F: float64(1.0),
									},
								},
								"stringValue": {
									TestValue: &protoMessages.Value_S{
										S: "string",
									},
								},
								"boolValue": {
									TestValue: &protoMessages.Value_B{
										B: true,
									},
								},
								"sliceValue": {
									TestValue: &protoMessages.Value_A{
										A: &protoMessages.ArrayValue{
											Items: []*protoMessages.Value{
												{
													TestValue: &protoMessages.Value_I{
														I: int64(1),
													},
												},
												{
													TestValue: &protoMessages.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"nilSlicePtrValue": {
									TestValue: &protoMessages.Value_Nil{
										Nil: true,
									},
								},
								"emptySlice": {
									TestValue: &protoMessages.Value_A{
										A: &protoMessages.ArrayValue{
											Items: nil,
										},
									},
								},
								"arrayValue": {
									TestValue: &protoMessages.Value_A{
										A: &protoMessages.ArrayValue{
											Items: []*protoMessages.Value{
												{
													TestValue: &protoMessages.Value_I{
														I: int64(1),
													},
												},
												{
													TestValue: &protoMessages.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"bytesValue": {
									TestValue: &protoMessages.Value_Any{
										Any: []byte("somebytes"),
									},
								},
								"structValue": {
									TestValue: &protoMessages.Value_O{
										O: &protoMessages.ObjectValue{
											Props: map[string]*protoMessages.Value{
												"IntValue": {
													TestValue: &protoMessages.Value_I{
														I: int64(1),
													},
												},
												"StringValue": {
													TestValue: &protoMessages.Value_S{
														S: "string",
													},
												},
												"taggedValue": {
													TestValue: &protoMessages.Value_S{
														S: "tagged",
													},
												},
											},
										},
									},
								},
								"mapValue": {
									TestValue: &protoMessages.Value_O{
										O: &protoMessages.ObjectValue{
											Props: map[string]*protoMessages.Value{
												"intValue": {
													TestValue: &protoMessages.Value_I{
														I: int64(1),
													},
												},
												"stringValue": {
													TestValue: &protoMessages.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"nilMapPtrValue": {
									TestValue: &protoMessages.Value_Nil{
										Nil: true,
									},
								},
								"emptyMap": {
									TestValue: &protoMessages.Value_O{
										O: &protoMessages.ObjectValue{
											Props: nil,
										},
									},
								},
								"emptyIsMarshaled": {
									TestValue: &protoMessages.Value_Nil{
										Nil: true,
									},
								},
							},
						},
					},
				},
				Protocol: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				SubscriptionPayload: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
			},
			ProtoResponse: &protoMessages.FieldResolveResponse{
				Response: &protoMessages.Value{
					TestValue: &protoMessages.Value_S{
						S: "response",
					},
				},
			},
			Expected: driver.FieldResolveOutput{
				Response: "response",
			},
		},
		{
			Title: "MarshalingObjectFieldsSource",
			Input: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{},
				Source: []*ast.ObjectField{
					{
						Name: &ast.Name{
							Value: "intField",
						},
						Value: &ast.IntValue{
							Value: "1",
						},
					},
					{
						Name: &ast.Name{
							Value: "stringField",
						},
						Value: &ast.StringValue{
							Value: "string",
						},
					},
				},
			},
			ProtoRequest: &protoMessages.FieldResolveRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Info: &protoMessages.FieldResolveInfo{},
				Source: &protoMessages.Value{
					TestValue: &protoMessages.Value_O{
						O: &protoMessages.ObjectValue{
							Props: map[string]*protoMessages.Value{
								"intField": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"stringField": {
									TestValue: &protoMessages.Value_S{
										S: "string",
									},
								},
							},
						},
					},
				},
				Protocol: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				SubscriptionPayload: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
			},
			ProtoResponse: &protoMessages.FieldResolveResponse{
				Response: &protoMessages.Value{
					TestValue: &protoMessages.Value_S{
						S: "response",
					},
				},
			},
			Expected: driver.FieldResolveOutput{
				Response: "response",
			},
		},
		{
			Title: "UnmarshalingResponse",
			Input: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{},
			},
			ProtoRequest: &protoMessages.FieldResolveRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Info: &protoMessages.FieldResolveInfo{},
				Source: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				Protocol: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				SubscriptionPayload: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
			},
			ProtoResponse: &protoMessages.FieldResolveResponse{
				Response: &protoMessages.Value{
					TestValue: &protoMessages.Value_O{
						O: &protoMessages.ObjectValue{
							Props: map[string]*protoMessages.Value{
								"intValue": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"uintValue": {
									TestValue: &protoMessages.Value_U{
										U: uint64(1),
									},
								},
								"floatValue": {
									TestValue: &protoMessages.Value_F{
										F: float64(1.0),
									},
								},
								"stringValue": {
									TestValue: &protoMessages.Value_S{
										S: "string",
									},
								},
								"boolValue": {
									TestValue: &protoMessages.Value_B{
										B: true,
									},
								},
								"sliceValue": {
									TestValue: &protoMessages.Value_A{
										A: &protoMessages.ArrayValue{
											Items: []*protoMessages.Value{
												{
													TestValue: &protoMessages.Value_I{
														I: int64(1),
													},
												},
												{
													TestValue: &protoMessages.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"bytesValue": {
									TestValue: &protoMessages.Value_Any{
										Any: []byte("somebytes"),
									},
								},
								"mapValue": {
									TestValue: &protoMessages.Value_O{
										O: &protoMessages.ObjectValue{
											Props: map[string]*protoMessages.Value{
												"intValue": {
													TestValue: &protoMessages.Value_I{
														I: int64(1),
													},
												},
												"stringValue": {
													TestValue: &protoMessages.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"emptyIsMarshaled": {
									TestValue: &protoMessages.Value_Nil{
										Nil: true,
									},
								},
							},
						},
					},
				},
			},
			Expected: driver.FieldResolveOutput{
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
}

// RunFieldResolveClientTests runs all client tests on a function
func RunFieldResolveClientTests(t *testing.T, f func(t *testing.T, tt FieldResolveClientTest)) {
	for _, tt := range FieldResolveClientTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}

// FieldResolveServerTest is basic struct for testing servers implementing proto
type FieldResolveServerTest struct {
	Title           string
	Input           *protoMessages.FieldResolveRequest
	HandlerInput    driver.FieldResolveInput
	HandlerResponse interface{}
	HandlerError    error
	Expected        *protoMessages.FieldResolveResponse
}

// FieldResolveServerTestData is a data for testing field resolution of proto servers
func FieldResolveServerTestData() []FieldResolveServerTest {
	return []FieldResolveServerTest{
		{
			Title: "PassesCorrectSource",
			Input: &protoMessages.FieldResolveRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Info: &protoMessages.FieldResolveInfo{},
				Source: &protoMessages.Value{
					TestValue: &protoMessages.Value_O{
						O: &protoMessages.ObjectValue{
							Props: map[string]*protoMessages.Value{
								"intValue": {
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
								"uintValue": {
									TestValue: &protoMessages.Value_U{
										U: uint64(1),
									},
								},
								"floatValue": {
									TestValue: &protoMessages.Value_F{
										F: float64(1.0),
									},
								},
								"stringValue": {
									TestValue: &protoMessages.Value_S{
										S: "string",
									},
								},
								"boolValue": {
									TestValue: &protoMessages.Value_B{
										B: true,
									},
								},
								"sliceValue": {
									TestValue: &protoMessages.Value_A{
										A: &protoMessages.ArrayValue{
											Items: []*protoMessages.Value{
												{
													TestValue: &protoMessages.Value_I{
														I: int64(1),
													},
												},
												{
													TestValue: &protoMessages.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"emptySlice": {
									TestValue: &protoMessages.Value_A{
										A: &protoMessages.ArrayValue{
											Items: nil,
										},
									},
								},
								"bytesValue": {
									TestValue: &protoMessages.Value_Any{
										Any: []byte("somebytes"),
									},
								},
								"mapValue": {
									TestValue: &protoMessages.Value_O{
										O: &protoMessages.ObjectValue{
											Props: map[string]*protoMessages.Value{
												"intValue": {
													TestValue: &protoMessages.Value_I{
														I: int64(1),
													},
												},
												"stringValue": {
													TestValue: &protoMessages.Value_S{
														S: "string",
													},
												},
											},
										},
									},
								},
								"emptyMap": {
									TestValue: &protoMessages.Value_O{
										O: &protoMessages.ObjectValue{
											Props: nil,
										},
									},
								},
								"emptyIsMarshaled": {
									TestValue: &protoMessages.Value_Nil{
										Nil: true,
									},
								},
							},
						},
					},
				},
			},
			HandlerInput: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{},
				Source: map[string]interface{}{
					"intValue":    int64(1),
					"uintValue":   uint64(1),
					"floatValue":  float64(1.0),
					"stringValue": "string",
					"boolValue":   true,
					"sliceValue":  []interface{}{int64(1), "string"},
					"emptySlice":  ([]interface{})(nil),
					"bytesValue":  []byte("somebytes"),
					"mapValue": map[string]interface{}{
						"intValue":    int64(1),
						"stringValue": "string",
					},
					"emptyMap":         (map[string]interface{})(nil),
					"emptyIsMarshaled": nil,
				},
			},
			HandlerResponse: "response",
			Expected: &protoMessages.FieldResolveResponse{
				Response: &protoMessages.Value{
					TestValue: &protoMessages.Value_S{
						S: "response",
					},
				},
			},
		},
		{
			Title: "PassesCorrectInfoObject",
			Input: &protoMessages.FieldResolveRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Info: &protoMessages.FieldResolveInfo{
					FieldName: "field",
					Path: &protoMessages.ResponsePath{
						Key: &protoMessages.Value{
							TestValue: &protoMessages.Value_S{
								S: "field",
							},
						},
						Prev: &protoMessages.ResponsePath{
							Key: &protoMessages.Value{
								TestValue: &protoMessages.Value_S{
									S: "fieldPrev",
								},
							},
						},
					},
					ReturnType: &protoMessages.TypeRef{
						TestTyperef: &protoMessages.TypeRef_List{
							List: &protoMessages.TypeRef{
								TestTyperef: &protoMessages.TypeRef_NonNull{
									NonNull: &protoMessages.TypeRef{
										TestTyperef: &protoMessages.TypeRef_Name{
											Name: "String",
										},
									},
								},
							},
						},
					},
					ParentType: &protoMessages.TypeRef{
						TestTyperef: &protoMessages.TypeRef_Name{
							Name: "SomeType",
						},
					},
					Operation: &protoMessages.OperationDefinition{
						Operation: "query",
						Name:      "getFieldOfFieldPrev",
						Directives: []*protoMessages.Directive{
							{
								Name: "@somedir",
								Arguments: map[string]*protoMessages.Value{
									"arg": {
										TestValue: &protoMessages.Value_S{
											S: "value",
										},
									},
								},
							},
						},
						SelectionSet: []*protoMessages.Selection{
							{
								Name: "subfield",
								Arguments: map[string]*protoMessages.Value{
									"arg": {
										TestValue: &protoMessages.Value_S{
											S: "value",
										},
									},
								},
								Directives: []*protoMessages.Directive{
									{
										Name: "@somedir",
										Arguments: map[string]*protoMessages.Value{
											"arg": {
												TestValue: &protoMessages.Value_S{
													S: "value",
												},
											},
										},
									},
								},
								SelectionSet: []*protoMessages.Selection{
									{
										Definition: &protoMessages.FragmentDefinition{
											Directives: []*protoMessages.Directive{
												{
													Name: "@somedir",
													Arguments: map[string]*protoMessages.Value{
														"arg": {
															TestValue: &protoMessages.Value_S{
																S: "value",
															},
														},
													},
												},
											},
											TypeCondition: &protoMessages.TypeRef{
												TestTyperef: &protoMessages.TypeRef_Name{
													Name: "SomeType",
												},
											},
											SelectionSet: []*protoMessages.Selection{
												{
													Name: "someField",
												},
											},
											VariableDefinitions: []*protoMessages.VariableDefinition{
												{
													Variable: &protoMessages.Variable{
														Name: "variable",
													},
													DefaultValue: &protoMessages.Value{
														TestValue: &protoMessages.Value_S{
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
					VariableValues: map[string]*protoMessages.Value{
						"var": {
							TestValue: &protoMessages.Value_S{
								S: "value",
							},
						},
					},
				},
			},
			HandlerInput: driver.FieldResolveInput{
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
						Operation: "query",
						Name:      "getFieldOfFieldPrev",
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
												{
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
			HandlerResponse: "response",
			Expected: &protoMessages.FieldResolveResponse{
				Response: &protoMessages.Value{
					TestValue: &protoMessages.Value_S{
						S: "response",
					},
				},
			},
		},
		{
			Title: "HandlesIndexResponse",
			Input: &protoMessages.FieldResolveRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Info: &protoMessages.FieldResolveInfo{
					FieldName: "field",
					Path: &protoMessages.ResponsePath{
						Key: &protoMessages.Value{
							TestValue: &protoMessages.Value_S{
								S: "field",
							},
						},
						Prev: &protoMessages.ResponsePath{
							Key: &protoMessages.Value{
								TestValue: &protoMessages.Value_I{
									I: int64(1),
								},
							},
						},
					},
				},
			},
			HandlerInput: driver.FieldResolveInput{
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
			HandlerResponse: "response",
			Expected: &protoMessages.FieldResolveResponse{
				Response: &protoMessages.Value{
					TestValue: &protoMessages.Value_S{
						S: "response",
					},
				},
			},
		},
		{
			Title: "ServerReturnsVariableValue",
			Input: &protoMessages.FieldResolveRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Info: &protoMessages.FieldResolveInfo{
					FieldName: "field",
					Operation: &protoMessages.OperationDefinition{
						VariableDefinitions: []*protoMessages.VariableDefinition{
							{
								Variable: &protoMessages.Variable{
									Name: "someVar",
								},
								DefaultValue: &protoMessages.Value{
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
							},
							{
								Variable: &protoMessages.Variable{
									Name: "someOtherVar",
								},
								DefaultValue: &protoMessages.Value{
									TestValue: &protoMessages.Value_I{
										I: int64(1),
									},
								},
							},
						},
						Directives: []*protoMessages.Directive{
							{
								Arguments: map[string]*protoMessages.Value{
									"arg": {
										TestValue: &protoMessages.Value_Variable{
											Variable: "someVar",
										},
									},
									"arg2": {
										TestValue: &protoMessages.Value_Variable{
											Variable: "someOtherVar",
										},
									},
								},
							},
						},
					},
					VariableValues: map[string]*protoMessages.Value{
						"someOtherVar": {
							TestValue: &protoMessages.Value_I{
								I: int64(2),
							},
						},
					},
				},
			},
			HandlerInput: driver.FieldResolveInput{
				Function: types.Function{
					Name: "function",
				},
				Info: driver.FieldResolveInfo{
					Operation: &types.OperationDefinition{
						VariableDefinitions: []types.VariableDefinition{
							{
								Variable: types.Variable{
									Name: "someVar",
								},
								DefaultValue: int64(1),
							},
							{
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
			HandlerResponse: "response",
			Expected: &protoMessages.FieldResolveResponse{
				Response: &protoMessages.Value{
					TestValue: &protoMessages.Value_S{
						S: "response",
					},
				},
			},
		},
	}
}

// RunFieldResolveServerTests runs all client tests on a function
func RunFieldResolveServerTests(t *testing.T, f func(t *testing.T, tt FieldResolveServerTest)) {
	for _, tt := range FieldResolveServerTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}
