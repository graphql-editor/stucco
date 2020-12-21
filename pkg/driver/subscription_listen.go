package driver

import "github.com/graphql-editor/stucco/pkg/types"

// SubscriptionListenInput represents input to a function which listen on events that trigger subscription
type SubscriptionListenInput struct {
	Function       types.Function
	Query          string                 `json:"query,omitempty"`
	VariableValues map[string]interface{} `json:"variableValues,omitempty"`
	OperationName  string                 `json:"operationName,omitempty"`
	Protocol       interface{}            `json:"protocol,omitempty"`
}

// SubscriptionListenReader is a simple interface that listens for pings from backing function
type SubscriptionListenReader interface {
	// Error returns the status of subscription listener that is no longer available for reading, if there was no error and stream was properly closed, it returns nil.
	Error() error
	// Next is blocking call that returns true when a new subscription should be started or false when listener is finished.
	Next() bool
	// Close closes the reader
	Close() error
}

// SubscriptionListenOutput represents response from a function which listen on events that trigger subscription
type SubscriptionListenOutput struct {
	Error  *Error `json:"error,omitempty"`
	Reader SubscriptionListenReader
}
