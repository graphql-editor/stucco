package driver

import "github.com/graphql-editor/stucco/pkg/types"

type StreamInfo struct {
	FieldName      string                     `json:"fieldName"`
	Path           *types.ResponsePath        `json:"path,omitempty"`
	ReturnType     *types.TypeRef             `json:"returnType,omitempty"`
	ParentType     *types.TypeRef             `json:"parentType,omitempty"`
	Operation      *types.OperationDefinition `json:"operation,omitempty"`
	VariableValues map[string]interface{}     `json:"variableValues,omitempty"`
}

type StreamMessage struct {
	Response interface{} `json:"response,omitempty"`
	Error    *Error      `json:"error,"`
}

type StreamReader interface {
	// Error returns the status of stream that is no longer available for reading, if there was no error and stream was properly closed, it returns nil.
	Error() error
	// Next is blocking operation that waits until next message is available or until stream is no longer available for reading. When next message is available function returns true, otherwise it returns false.
	Next() bool
	// Read returns next message in stream. Read can only by called after Next that returned true.
	Read() StreamMessage
	// Close stream
	Close()
}

type StreamInput struct {
	Function  types.Function
	Arguments types.Arguments `json:"arguments,omitempty"`
	Info      StreamInfo      `json:"info"`
	Secrets   Secrets         `json:"secrets,omitempty"`
	Protocol  interface{}     `json:"protocol,omitempty"`
}

type StreamOutput struct {
	Reader StreamReader
}
