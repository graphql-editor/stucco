package deployment

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment/config"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2019-08-01/web"
)

// Option used to create deployment context
type Option interface {
	withOpt(opts *options)
}

type options struct {
	projectName                 string
	blobContainer               string
	resourceGroup               string
	resourceGroupLocation       string
	storageAccount              string
	schema                      string
	stuccoJSON                  string
	functionAppName             string
	functionAppPlanName         string
	functionAppLocation         string
	functionAppPlanSku          string
	functionAppPlanWorkers      int32
	functionAppImage            FunctionAppImage
	secrets                     map[string]string
	blobContainerSource         BlobContainerSource
	resourceGroupSource         ResourceGroupSource
	storageAccountSource        StorageAccountSource
	schemaSource                SchemaSource
	stuccoJSONSource            StuccoJSONSource
	resourceGroupLocationSource LocationSource
	functionAppSource           FunctionAppSource
	secretsSource               SecretsSource
}

// Context represents data gathered during the deployment of stucco functions to Azure Functions
type Context struct {
	config.Config  `json:"-"`
	BlobContainer  BlobContainer     `json:"blobContainer"`
	ResourceGroup  ResourceGroup     `json:"resourceGroup"`
	StorageAccount StorageAccount    `json:"storageAccount"`
	FunctionApp    FunctionApp       `json:"functionApp"`
	Schema         string            `json:"-"`
	StuccoJSON     string            `json:"-"`
	Secrets        map[string]string `json:"-"`
}

// Builder creates context object from options
type Builder struct {
	config.Config

	opts options
}

func (b *Builder) prepareResourceGroup(ctx context.Context, c *Context) (err error) {
	if err = ctx.Err(); err == nil {
		groupsClient := resources.NewGroupsClient(b.SubscriptionID)
		groupsClient.Authorizer = b.Authorizer
		c.ResourceGroup, err = b.opts.resourceGroupSource.Select(ctx, groupsClient)
	}
	if err == nil {
		err = ctx.Err()
	}
	if err == nil {
		c.ResourceGroup, err = prepareResourceGroup(ctx, c.Config, c.ResourceGroup, b.opts.resourceGroupLocationSource)
	}
	return
}

func (b *Builder) prepareStorageAccount(ctx context.Context, c *Context) (err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	accountsClient := storage.NewAccountsClient(b.SubscriptionID)
	accountsClient.Authorizer = b.Authorizer
	c.StorageAccount, err = b.opts.storageAccountSource.Select(
		ctx,
		accountsClient,
		*c,
	)
	if err == nil {
		err = ctx.Err()
	}
	if err == nil {
		c.StorageAccount, err = prepareStorageAccount(
			ctx,
			accountsClient,
			*c,
		)
	}
	return
}

func (b *Builder) prepareBlobContainer(ctx context.Context, c *Context) (err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	blobContainersClient := storage.NewBlobContainersClient(b.SubscriptionID)
	blobContainersClient.Authorizer = b.Authorizer
	c.BlobContainer, err = b.opts.blobContainerSource.Select(
		ctx,
		blobContainersClient,
		*c,
	)
	if err == nil {
		err = ctx.Err()
	}
	if err == nil {
		c.BlobContainer, err = prepareBlobContainer(
			ctx,
			blobContainersClient,
			*c,
		)
	}
	return
}

func (b *Builder) prepareFunctionAppSecrets(ctx context.Context, c *Context) (err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	c.Secrets, err = b.opts.secretsSource.Select(
		ctx,
		*c,
	)
	if err == nil {
		err = ctx.Err()
	}
	return
}

func (b *Builder) prepareFunctionApp(ctx context.Context, c *Context) (err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	client := web.New(b.SubscriptionID)
	client.Authorizer = b.Authorizer
	c.FunctionApp, err = b.opts.functionAppSource.Select(
		ctx,
		*c,
		client,
	)
	if err == nil {
		err = ctx.Err()
	}
	if err == nil {
		c.FunctionApp, err = prepareFunctionApp(
			ctx,
			*c,
		)
	}
	return
}

// NewBuilder creates new context of deployment
func NewBuilder(c config.Config, opts ...Option) *Builder {
	b := Builder{Config: c}
	b.opts = options{
		blobContainerSource: defaultBlobContainerSource{
			opts: &b.opts,
		},
		resourceGroupSource: defaultResourceGroupSource{
			opts: &b.opts,
		},
		storageAccountSource: defaultStorageAccountSource{
			opts: &b.opts,
		},
		schemaSource: defaultSchemaSource{
			opts: &b.opts,
		},
		stuccoJSONSource: defaultStuccoJSONSource{
			opts: &b.opts,
		},
		resourceGroupLocationSource: defaultResourceGroupLocationSource{
			opts: &b.opts,
		},
		functionAppSource: defaultFunctionAppSource{
			opts: &b.opts,
		},
	}
	for _, opt := range opts {
		opt.withOpt(&b.opts)
	}
	return &b
}

// BuildContext creates new context using builder
func BuildContext(ctx context.Context, b *Builder) (c Context, err error) {
	c.Config = b.Config
	err = b.prepareResourceGroup(ctx, &c)
	if err == nil {
		err = b.prepareStorageAccount(ctx, &c)
	}
	if err == nil {
		err = b.prepareBlobContainer(ctx, &c)
	}
	if err == nil {
		err = b.prepareFunctionApp(ctx, &c)
	}
	if err == nil {
		err = b.prepareFunctionAppSecrets(ctx, &c)
	}
	c.Schema = b.opts.schema
	c.StuccoJSON = b.opts.stuccoJSON
	return
}
