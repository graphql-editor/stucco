package prompts

import (
	"context"
	"strings"

	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
)

// StorageAccountSource for client runs
type StorageAccountSource string

// Select creates new name for storage account
// If it's set, the value is used, otherwise it fallsback to stuccosa
func (s StorageAccountSource) Select(
	ctx context.Context,
	client storage.AccountsClient,
	context deployment.Context,
) (sa deployment.StorageAccount, err error) {
	if string(s) != "" {
		sa = deployment.StorageAccount(s)
		return
	}
	name := strings.Replace(context.ResourceGroup.Name, "-", "", -1)
	if len(name) > 24 {
		name = name[:24]
	}
	sa = deployment.StorageAccount(name)
	return
}
