package driver

import "github.com/graphql-editor/stucco/pkg/types"

// SubscriptionConnectionInput represents input to a function which creates subscription connection data
type SubscriptionConnectionInput struct {
	Function       types.Function
	Query          string                 `json:"query,omitempty"`
	VariableValues map[string]interface{} `json:"variableValues,omitempty"`
	OperationName  string                 `json:"operationName,omitempty"`
	Protocol       interface{}            `json:"protocol,omitempty"`
}

// SubscriptionConnectionOutput represents response from a function which creates subscription connection data
type SubscriptionConnectionOutput struct {
	Response interface{} `json:"response,omitempty"`
	Error    *Error      `json:"error,omitempty"`
}
