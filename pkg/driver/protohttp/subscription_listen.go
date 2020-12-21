package protohttp

import "github.com/graphql-editor/stucco/pkg/driver"

// SubscriptionListen implements driver.SubscriptionListen. Currently protocol buffer subscription listening is not supported
// over HTTP
func (c *Client) SubscriptionListen(driver.SubscriptionListenInput) driver.SubscriptionListenOutput {
	return driver.SubscriptionListenOutput{
		Error: &driver.Error{
			Message: "HTTP transport does not subscription listening. Try using external subscription",
		},
	}
}
