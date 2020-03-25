package deployment

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment/config"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-06-01/subscriptions"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/pkg/errors"
)

func prepareResourceGroup(
	ctx context.Context,
	c config.Config,
	inGroup ResourceGroup,
	source LocationSource,
) (outGroup ResourceGroup, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	outGroup = inGroup
	if inGroup.Location == "" {
		var locations []subscriptions.Location
		if locations, err = getAvailableLocations(c); err == nil {
			var location Location
			if location, err = source.Select(ctx, locations); err == nil {
				outGroup.Location = location
			}
		}
	}
	return
}

func prepareStorageAccount(
	ctx context.Context,
	client storage.AccountsClient,
	buildContext Context,
) (account StorageAccount, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	account = buildContext.StorageAccount
	return
}

func prepareBlobContainer(
	ctx context.Context,
	client storage.BlobContainersClient,
	buildContext Context,
) (blob BlobContainer, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	blob = buildContext.BlobContainer
	return
}

// ValidFunctionAppLocation checks if current location is in one of supported function
// app locations
func ValidFunctionAppLocation(fp FunctionApp) string {
	for _, loc := range FunctionLocations {
		if string(fp.Location) == loc.Name || string(fp.Location) == loc.Slug {
			return loc.Slug
		}
	}
	return ""
}

// Plan for Azure
type Plan struct {
	Sku     string
	Display string
}

// SupportedPlans is a list of plans on which stucco can be deployed
var SupportedPlans = []Plan{
	{
		Sku:     "F1",
		Display: "Free 1 App Service plan - dev only",
	}, {
		Sku:     "B1",
		Display: "Basic 1 App Service plan",
	}, {
		Sku:     "B2",
		Display: "Basic 2 App Service plan",
	}, {
		Sku:     "B3",
		Display: "Basic 3 App Service plan",
	}, {
		Sku:     "S1",
		Display: "Standard 1 App Service plan",
	}, {
		Sku:     "S2",
		Display: "Standard 2 App Service plan",
	}, {
		Sku:     "S3",
		Display: "Standard 3 App Service plan",
	}, {
		Sku:     "P1V2",
		Display: "Premium V2 1 App Service plan",
	}, {
		Sku:     "P2V2",
		Display: "Premium V2 2 App Service plan",
	}, {
		Sku:     "P3V3",
		Display: "Premium V2 3 App Service plan",
	}, {
		Sku:     "EP1",
		Display: "Elastic Premium 1 plan",
	}, {
		Sku:     "EP2",
		Display: "Elastic Premium 2 plan",
	},
}

// ValidFunctionAppSku checks if current location is in one of supported function
// app skus
func ValidFunctionAppSku(fp FunctionAppPlan) string {
	for _, plan := range SupportedPlans {
		if fp.Sku == plan.Sku {
			return plan.Sku
		}
	}
	return ""
}

func prepareFunctionApp(
	ctx context.Context,
	buildContext Context,
) (fp FunctionApp, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	fp = buildContext.FunctionApp
	if fp.Location == "" {
		fp.Location = buildContext.ResourceGroup.Location
	}
	if ValidFunctionAppLocation(fp) == "" {
		err = errors.Errorf("Location '%s' is not a valid location for Function App", string(fp.Location))
	}
	if err == nil && ValidFunctionAppSku(fp.Plan) == "" {
		err = errors.Errorf("SKU '%s' is not a valid sku for Function App", string(fp.Plan.Sku))
	}
	return
}
