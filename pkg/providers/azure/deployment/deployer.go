package deployment

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/graphql-editor/stucco/pkg/printer"
	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment/config"
	"github.com/graphql-editor/stucco/pkg/providers/azure/driver"
	"github.com/graphql-editor/stucco/pkg/version"

	"github.com/AlekSi/pointer"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2019-08-01/web"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest"
	"github.com/pkg/errors"
)

var (
	stuccoRouterImage = "gqleditor/stucco-router-azure-worker:" + version.Version
)

func isDetailedError404(err error) (autorest.DetailedError, bool) {
	detailedError, ok := err.(autorest.DetailedError)
	ok = ok && detailedError.StatusCode == http.StatusNotFound
	return detailedError, ok
}

// DeployClient deploys user settings according to context
type DeployClient struct {
	config.Config
}

type deployContext struct {
	config.Config
	group              resources.Group
	storageAccountKeys []storage.AccountKey
	blobContainer      storage.BlobContainer
	appServicePlan     web.AppServicePlan
	routerFunction     web.Site
	userFunction       web.Site
	functionHostKeys   web.HostKeys
}

func (d *deployContext) deployGroup(ctx context.Context, context Context) (err error) {
	printer.ColorPrintf("Preparing resource group %s ...\n", context.ResourceGroup.Name)
	groupsClient := resources.NewGroupsClient(d.SubscriptionID)
	groupsClient.Authorizer = d.Authorizer
	d.group, err = groupsClient.CreateOrUpdate(ctx, context.ResourceGroup.Name, resources.Group{
		Location: (*string)(&context.ResourceGroup.Location),
		Tags: map[string]*string{
			"managedBy": pointer.ToString("stucco"),
		},
	})
	return
}

func (d *deployContext) getOrCreateBlobContainer(ctx context.Context, context Context, container string) (c storage.BlobContainer, created bool, err error) {
	blobContainersClient := storage.NewBlobContainersClient(d.SubscriptionID)
	blobContainersClient.Authorizer = d.Authorizer
	c, err = blobContainersClient.Get(ctx, context.ResourceGroup.Name, string(context.StorageAccount), container)
	if _, ok := isDetailedError404(err); ok {
		created = true
		c, err = blobContainersClient.Create(
			ctx,
			context.ResourceGroup.Name,
			string(context.StorageAccount),
			container,
			storage.BlobContainer{
				ContainerProperties: &storage.ContainerProperties{
					PublicAccess: storage.PublicAccessNone,
				},
			},
		)
	}
	return
}

func (d *deployContext) deployFunctionKeys(ctx context.Context, context Context) (err error) {
	if _, _, err := d.getOrCreateBlobContainer(ctx, context, "azure-webjobs-secrets"); err == nil {
		k := context.FunctionApp.Key
		if k == "" {
			b := make([]byte, 32)
			_, err = rand.Read(b)
			k = base64.RawURLEncoding.EncodeToString(b)
		}
		var buf bytes.Buffer
		if err == nil {
			err = json.NewEncoder(&buf).Encode(driver.HostJSON{
				MasterKey: driver.HostJSONKey{
					Name:  "masterKey",
					Value: k,
				},
				FunctionKeys: make([]driver.HostJSONKey, 0),
			})
		}
		if err == nil {
			err = d.putBlobIn(
				ctx,
				context.FunctionApp.Name+"-functions/host.json",
				buf.String(),
				"application/json",
				context,
				"azure-webjobs-secrets",
			)
		}
	}
	return
}

func (d *deployContext) deployStorageAccount(ctx context.Context, context Context) (err error) {
	printer.ColorPrintf("Preparing storage account %s ...\n", string(context.StorageAccount))
	accountsClient := storage.NewAccountsClient(d.SubscriptionID)
	accountsClient.Authorizer = d.Authorizer
	var keyList storage.AccountListKeysResult
	keyList, err = accountsClient.ListKeys(ctx, context.ResourceGroup.Name, string(context.StorageAccount), "")
	// If no account exists, create it, otherwise reuse existing one.
	if _, ok := isDetailedError404(err); ok {
		var accountFuture storage.AccountsCreateFuture
		accountFuture, err = accountsClient.Create(ctx, context.ResourceGroup.Name, string(context.StorageAccount), storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardRAGRS,
				Tier: storage.Standard,
			},
			Kind:     storage.BlobStorage,
			Location: (*string)(&context.ResourceGroup.Location),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{
				// Is cool enough?
				AccessTier: storage.Hot,
				NetworkRuleSet: &storage.NetworkRuleSet{
					DefaultAction: storage.DefaultActionAllow,
				},
			},
		})
		if err == nil {
			err = accountFuture.WaitForCompletionRef(ctx, accountsClient.Client)
		}
		if err == nil {
			_, err = accountFuture.Result(accountsClient)
		}
		if err == nil {
			keyList, err = accountsClient.ListKeys(ctx, context.ResourceGroup.Name, string(context.StorageAccount), "")
		}
	}
	if err == nil && keyList.Keys != nil {
		d.storageAccountKeys = append(d.storageAccountKeys, (*keyList.Keys)...)
	}
	if err == nil {
		err = d.deployFunctionKeys(ctx, context)
	}
	return
}

func (d *deployContext) blobCredential(context Context) (*azblob.SharedKeyCredential, error) {
	accountName := string(context.StorageAccount)
	accountKey := *d.storageAccountKeys[0].Value
	return azblob.NewSharedKeyCredential(accountName, accountKey)
}

func (d *deployContext) blobURL(file, sa, container string, context Context) (u azblob.BlockBlobURL, err error) {
	var credential *azblob.SharedKeyCredential
	credential, err = d.blobCredential(context)
	var cURL *url.URL
	if err == nil {
		cURL, err = url.Parse(fmt.Sprintf(
			"https://%s.blob.core.windows.net/%s",
			sa,
			container,
		))
	}
	if err == nil {
		p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
		containerURL := azblob.NewContainerURL(*cURL, p)
		u = containerURL.NewBlockBlobURL(file)
	}
	return
}

func (d *deployContext) putBlobIn(
	ctx context.Context,
	name,
	content,
	contentType string,
	context Context,
	container string,
) error {
	u, err := d.blobURL(name, string(context.StorageAccount), container, context)
	if err == nil {
		rd := strings.NewReader(content)
		_, err = u.Upload(ctx,
			rd,
			azblob.BlobHTTPHeaders{
				ContentType: contentType,
			},
			azblob.Metadata{},
			azblob.BlobAccessConditions{},
		)
	}
	return err
}

func (d *deployContext) putBlob(
	ctx context.Context,
	name,
	content,
	contentType string,
	context Context,
) error {
	return d.putBlobIn(
		ctx,
		name,
		content,
		contentType,
		context,
		string(context.BlobContainer),
	)
}

func (d *deployContext) blobSignatureURL(file string, context Context) (sig string, err error) {
	var credential *azblob.SharedKeyCredential
	credential, err = d.blobCredential(context)
	var queryParams azblob.SASQueryParameters
	if err == nil {
		queryParams, err = azblob.BlobSASSignatureValues{
			Protocol:      azblob.SASProtocolHTTPS,
			ExpiryTime:    time.Now().Add(time.Hour * 24 * 365 * 10), // 10 years
			ContainerName: string(context.BlobContainer),
			BlobName:      file,
			Permissions:   azblob.BlobSASPermissions{Read: true}.String(),
		}.NewSASQueryParameters(credential)
	}
	if err == nil {
		sig = fmt.Sprintf(
			"https://%s.blob.core.windows.net/%s/%s?%s",
			string(context.StorageAccount),
			string(context.BlobContainer),
			file,
			queryParams.Encode(),
		)
	}
	return
}

func (d *deployContext) deployBlobStorageWithFiles(ctx context.Context, context Context) (err error) {
	printer.ColorPrintf("Preparing blob storage %s ...\n", string(context.StorageAccount))
	d.blobContainer, _, err = d.getOrCreateBlobContainer(ctx, context, string(context.BlobContainer))
	if err == nil && context.Schema != "" {
		err = d.putBlob(ctx, "schema.graphql", context.Schema, "application/graphql", context)
	}
	if err == nil && context.StuccoJSON != "" {
		err = d.putBlob(ctx, "stucco.json", context.StuccoJSON, "application/json", context)
	}
	return
}

func (d *deployContext) deployApplicationPlan(ctx context.Context, context Context) (err error) {
	appServicePlansClient := web.NewAppServicePlansClient(d.SubscriptionID)
	appServicePlansClient.Authorizer = d.Authorizer
	appServicePlanFuture, err := appServicePlansClient.CreateOrUpdate(ctx, context.ResourceGroup.Name, context.FunctionApp.Plan.Name, web.AppServicePlan{
		Sku: &web.SkuDescription{
			Name:     &context.FunctionApp.Plan.Sku,
			Capacity: &context.FunctionApp.Plan.Workers,
		},
		AppServicePlanProperties: &web.AppServicePlanProperties{
			PerSiteScaling: pointer.ToBool(false),
			IsXenon:        pointer.ToBool(false),
			Reserved:       pointer.ToBool(true),
		},
		Location: (*string)(&context.FunctionApp.Location),
	})
	if err == nil {
		err = appServicePlanFuture.WaitForCompletionRef(ctx, appServicePlansClient.Client)
	}
	if err == nil {
		d.appServicePlan, err = appServicePlanFuture.Result(appServicePlansClient)
	}
	return
}

func (d *deployContext) newFunctionSite(image FunctionAppImage, context Context, appSettings []web.NameValuePair) web.Site {
	appSettings = append([]web.NameValuePair{
		{
			Name:  pointer.ToString("DOCKER_CUSTOM_IMAGE_NAME"),
			Value: &image.Repository,
		},
		{
			Name:  pointer.ToString("FUNCTION_APP_EDIT_MODE"),
			Value: pointer.ToString("readOnly"),
		},
		{
			Name:  pointer.ToString("WEBSITES_ENABLE_APP_SERVICE_STORAGE"),
			Value: pointer.ToString("false"),
		},
		{
			Name:  pointer.ToString("FUNCTIONS_EXTENSION_VERSION"),
			Value: pointer.ToString("~3"),
		},
		{
			Name: pointer.ToString("AzureWebJobsStorage"),
			Value: pointer.ToString(fmt.Sprintf(
				"DefaultEndpointsProtocol=https;EndpointSuffix=core.windows.net;AccountName=%s;AccountKey=%s",
				string(context.StorageAccount),
				*d.storageAccountKeys[0].Value,
			)),
		},
		{
			Name: pointer.ToString("AzureWebJobsDashboard"),
			Value: pointer.ToString(fmt.Sprintf(
				"DefaultEndpointsProtocol=https;EndpointSuffix=core.windows.net;AccountName=%s;AccountKey=%s",
				string(context.StorageAccount),
				*d.storageAccountKeys[0].Value,
			)),
		},
		{
			Name:  pointer.ToString("WEBSITE_NODE_DEFAULT_VERSION"),
			Value: pointer.ToString("~12"),
		},
		// Timestamp so that each new update is "touched", this is needed
		// as each update rotates host master key and functions need
		// to be restared for that to change to take place.
		{
			Name:  pointer.ToString("UPDATED_AT"),
			Value: pointer.ToString(time.Now().String()),
		},
	}, appSettings...)
	if image.Registry != "" {
		appSettings = append(appSettings, web.NameValuePair{
			Name:  pointer.ToString("DOCKER_REGISTRY_SERVER_URL"),
			Value: &image.Registry,
		})
	}
	if image.Username != "" {
		appSettings = append(appSettings, web.NameValuePair{
			Name:  pointer.ToString("DOCKER_REGISTRY_SERVER_USERNAME"),
			Value: &image.Username,
		})
	}
	if image.Password != "" {
		appSettings = append(appSettings, web.NameValuePair{
			Name:  pointer.ToString("DOCKER_REGISTRY_SERVER_PASSWORD"),
			Value: &image.Password,
		})
	}
	return web.Site{
		Kind:     pointer.ToString("functionapp,linux,container"),
		Location: (*string)(&context.FunctionApp.Location),
		SiteProperties: &web.SiteProperties{
			ServerFarmID: &context.FunctionApp.Plan.Name,
			Reserved:     pointer.ToBool(false),
			IsXenon:      pointer.ToBool(false),
			HyperV:       pointer.ToBool(false),
			SiteConfig: &web.SiteConfig{
				NumberOfWorkers:     &context.FunctionApp.Plan.Workers,
				NetFrameworkVersion: pointer.ToString("v4.0"),
				LinuxFxVersion:      pointer.ToString(fmt.Sprintf("DOCKER|%s", image.Repository)),
				AppSettings:         &appSettings,
				AlwaysOn:            pointer.ToBool(false),
				LocalMySQLEnabled:   pointer.ToBool(false),
				HTTP20Enabled:       pointer.ToBool(true),
			},
			ScmSiteAlsoStopped: pointer.ToBool(false),
		},
		Tags: map[string]*string{
			"managedBy": pointer.ToString("stucco"),
			"apiName":   &context.FunctionApp.Name,
		},
	}
}

func (d *deployContext) deploySite(
	ctx context.Context,
	context Context,
	siteName string,
	site web.Site,
) (rt web.Site, err error) {
	appsClient := web.NewAppsClient(d.SubscriptionID)
	appsClient.Authorizer = d.Authorizer
	appFuture, err := appsClient.CreateOrUpdate(ctx, context.ResourceGroup.Name, siteName, site)
	if err == nil {
		err = appFuture.WaitForCompletionRef(ctx, appsClient.Client)
	}
	if err == nil {
		rt, err = appFuture.Result(appsClient)
	}
	return
}

func (d *deployContext) deployRouter(ctx context.Context, context Context) (err error) {
	printer.ColorPrintf("Preparing router function. This may take a moment ...\n")
	schemaURL, err := d.blobSignatureURL("schema.graphql", context)
	var stuccoJSONURL string
	if err == nil {
		stuccoJSONURL, err = d.blobSignatureURL("stucco.json", context)
	}
	var functionURL string
	if err == nil {
		functionURL, err = d.functionURL()
	}
	var rt web.Site
	if err == nil {
		rt, err = d.deploySite(ctx, context, context.FunctionApp.Name, d.newFunctionSite(FunctionAppImage{Repository: stuccoRouterImage}, context, []web.NameValuePair{
			{
				Name:  pointer.ToString("STUCCO_SCHEMA"),
				Value: &schemaURL,
			},
			{
				Name:  pointer.ToString("STUCCO_CONFIG"),
				Value: &stuccoJSONURL,
			},
			{
				Name:  pointer.ToString("STUCCO_AZURE_WORKER_BASE_URL"),
				Value: &functionURL,
			},
		}))
	}
	if err == nil {
		d.routerFunction = rt
	}
	return
}

func (d *deployContext) functionURL() (string, error) {
	if d.userFunction.SiteProperties.HostNames == nil || len(*d.userFunction.SiteProperties.HostNames) == 0 {
		return "", errors.Errorf("could not find host for site %s", pointer.GetString(d.userFunction.Name))
	}
	return fmt.Sprintf("https://%s", (*d.userFunction.SiteProperties.HostNames)[0]), nil
}

func (d *deployContext) routerURL() (string, error) {
	if d.routerFunction.SiteProperties.HostNames == nil || len(*d.routerFunction.SiteProperties.HostNames) == 0 {
		return "", errors.Errorf("could not find host for site %s", pointer.GetString(d.routerFunction.Name))
	}
	return fmt.Sprintf("https://%s", (*d.routerFunction.SiteProperties.HostNames)[0]), nil
}

func (d *deployContext) deployFunction(ctx context.Context, context Context) (err error) {
	printer.ColorPrintf("Preparing worker function. This may take a moment ...\n")
	nvp := []web.NameValuePair{}
	for k, v := range context.Secrets {
		key := k
		value := v
		nvp = append(nvp, web.NameValuePair{
			Name:  &key,
			Value: &value,
		})
	}
	rt, err := d.deploySite(ctx, context, context.FunctionApp.Name+"-functions", d.newFunctionSite(context.FunctionApp.Image, context, nvp))
	if err == nil {
		d.userFunction = rt
	}
	return
}

func (d *deployContext) deploy(ctx context.Context, context Context) (err error) {
	err = d.deployGroup(ctx, context)
	if err == nil {
		err = d.deployStorageAccount(ctx, context)
	}
	if err == nil {
		err = d.deployBlobStorageWithFiles(ctx, context)
	}
	printer.ColorPrintf("Preparing application environment. This may take a moment ...\n")
	if err == nil {
		err = d.deployApplicationPlan(ctx, context)
	}
	if err == nil {
		err = d.deployFunction(ctx, context)
	}
	if err == nil {
		err = d.deployRouter(ctx, context)
	}
	var routerHost string
	if err == nil {
		routerHost, err = d.routerURL()
	}
	if err == nil {
		printer.ColorPrintf("Done. Your GraphQL API should be available at %s/graphql in a few moments.\n", routerHost)
	}
	return
}

// Deploy context to Azure
func (d DeployClient) Deploy(ctx context.Context, context Context) (err error) {
	dctx := deployContext{Config: d.Config}
	return dctx.deploy(ctx, context)
}
