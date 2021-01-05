package driver

import "github.com/graphql-editor/stucco/pkg/types"

// FieldResolveInfo defines information about current field resolution
type FieldResolveInfo struct {
	FieldName      string                     `json:"fieldName"`
	Path           *types.ResponsePath        `json:"path,omitempty"`
	ReturnType     *types.TypeRef             `json:"returnType,omitempty"`
	ParentType     *types.TypeRef             `json:"parentType,omitempty"`
	Operation      *types.OperationDefinition `json:"operation,omitempty"`
	VariableValues map[string]interface{}     `json:"variableValues,omitempty"`
	RootValue      interface{}                `json:"rootValue,omitempty"`
}

// FieldResolveInput represents data passed to field resolution
type FieldResolveInput struct {
	Function  types.Function
	Source    interface{}      `json:"source,omitempty"`
	Arguments types.Arguments  `json:"arguments,omitempty"`
	Info      FieldResolveInfo `json:"info"`
	Protocol  interface{}      `json:"protocol,omitempty"`
}

// FieldResolveOutput is a result of a field resolution
type FieldResolveOutput struct {
	Response interface{} `json:"response,omitempty"`
	Error    *Error      `json:"error,omitempty"`
}
