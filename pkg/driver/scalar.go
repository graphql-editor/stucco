package driver

import "github.com/graphql-editor/stucco/pkg/types"

type ScalarParseInput struct {
	Function types.Function `json:"function"`
	Value    interface{}    `json:"value"`
}
type ScalarParseOutput struct {
	Response interface{} `json:"response,omitempty"`
	Error    *Error      `json:"error,omitempty"`
}

type ScalarSerializeInput struct {
	Function types.Function `json:"function"`
	Value    interface{}    `json:"value"`
}
type ScalarSerializeOutput struct {
	Response interface{} `json:"response,omitempty"`
	Error    *Error      `json:"error,omitempty"`
}
