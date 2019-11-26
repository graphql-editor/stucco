package driver

import "github.com/graphql-editor/stucco/pkg/types"

type InterfaceResolveTypeInfo struct {
	FieldName      string                     `json:"fieldName"`
	Path           *types.ResponsePath        `json:"path,omitempty"`
	ReturnType     *types.TypeRef             `json:"returnType,omitempty"`
	ParentType     *types.TypeRef             `json:"parentType,omitempty"`
	Operation      *types.OperationDefinition `json:"operation,omitempty"`
	VariableValues map[string]interface{}     `json:"variableValues,omitempty"`
}

type InterfaceResolveTypeInput struct {
	Function types.Function
	Value    interface{}
	Info     InterfaceResolveTypeInfo
}
type InterfaceResolveTypeOutput struct {
	Type  types.TypeRef
	Error *Error
}
