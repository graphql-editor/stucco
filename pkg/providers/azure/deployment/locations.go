package deployment

import (
	"context"
	"sync"

	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment/config"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-06-01/subscriptions"
)

var (
	locationsCache     = make(map[string][]subscriptions.Location)
	locationsCacheLock sync.Mutex
)

func getAvailableLocations(c config.Config) (locations []subscriptions.Location, err error) {
	locationsCacheLock.Lock()
	locations = locationsCache[c.SubscriptionID]
	locationsCacheLock.Unlock()
	if len(locations) > 0 {
		return
	}
	subscriptionsClient := subscriptions.NewClient()
	subscriptionsClient.Authorizer = c.Authorizer
	res, err := subscriptionsClient.ListLocations(context.Background(), c.SubscriptionID)
	if err != nil {
		return
	}
	if res.Value != nil {
		locations = *res.Value
	}
	locationsCacheLock.Lock()
	locationsCache[c.SubscriptionID] = locations
	locationsCacheLock.Unlock()
	return
}
