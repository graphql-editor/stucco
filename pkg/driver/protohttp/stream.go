package protohttp

import "github.com/graphql-editor/stucco/pkg/driver"

// Stream implements driver.Stream. Currently protocol buffer streaming is not supported
// over HTTP
func (c *Client) Stream(driver.StreamInput) driver.StreamOutput {
	return driver.StreamOutput{
		Error: &driver.Error{
			Message: "HTTP transport does not support streaming",
		},
	}
}
