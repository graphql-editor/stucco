package prompts

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
)

// BlobContainerSource for client runs
type BlobContainerSource string

// Select creates new name for blob container
// If it's set, the value is used, otherwise it fallsback to the name of resource group with suffix
func (b BlobContainerSource) Select(
	ctx context.Context,
	client storage.BlobContainersClient,
	context deployment.Context,
) (bc deployment.BlobContainer, err error) {
	if string(b) != "" {
		bc = deployment.BlobContainer(b)
		return
	}
	bc = deployment.BlobContainer(context.ResourceGroup.Name + "-stucco-files")
	return
}
