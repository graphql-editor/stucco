package driver

import "github.com/graphql-editor/stucco/pkg/types"

type UnionResolveTypeInfo struct {
	FieldName      string                     `json:"fieldName"`
	Path           *types.ResponsePath        `json:"path,omitempty"`
	ReturnType     *types.TypeRef             `json:"returnType,omitempty"`
	ParentType     *types.TypeRef             `json:"parentType,omitempty"`
	Operation      *types.OperationDefinition `json:"operation,omitempty"`
	VariableValues map[string]interface{}     `json:"variableValues,omitempty"`
}

type UnionResolveTypeInput struct {
	Function types.Function
	Value    interface{}
	Info     UnionResolveTypeInfo
}
type UnionResolveTypeOutput struct {
	Type  types.TypeRef
	Error *Error
}
