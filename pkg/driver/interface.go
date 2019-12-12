package driver

import "github.com/graphql-editor/stucco/pkg/types"

// InterfaceResolveTypeInfo contains information about current state of query
// for interface type resolution
type InterfaceResolveTypeInfo struct {
	FieldName      string                     `json:"fieldName"`
	Path           *types.ResponsePath        `json:"path,omitempty"`
	ReturnType     *types.TypeRef             `json:"returnType,omitempty"`
	ParentType     *types.TypeRef             `json:"parentType,omitempty"`
	Operation      *types.OperationDefinition `json:"operation,omitempty"`
	VariableValues map[string]interface{}     `json:"variableValues,omitempty"`
}

// InterfaceResolveTypeInput represents a request of interface type resolution for
// GraphQL query
type InterfaceResolveTypeInput struct {
	Function types.Function
	Value    interface{}
	Info     InterfaceResolveTypeInfo
}

// InterfaceResolveTypeOutput represents an output returned by runner for request of
// interface type resolution
type InterfaceResolveTypeOutput struct {
	Type  types.TypeRef
	Error *Error
}
