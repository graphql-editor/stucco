package driver

import "github.com/graphql-editor/stucco/pkg/types"

// AuthorizeInput represents data passed to authorize function
type AuthorizeInput struct {
	Function       types.Function         `json:"function,omitempty"`
	Query          string                 `json:"query,omitempty"`
	OperationName  string                 `json:"operationName,omitempty"`
	VariableValues map[string]interface{} `json:"variableValues,omitempty"`
	Protocol       interface{}            `json:"protocol,omitempty"`
}

// AuthorizeOutput is an authorize response
type AuthorizeOutput struct {
	Response bool   `json:"response,omitempty"`
	Error    *Error `json:"error,omitempty"`
}
