package deployment

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-06-01/subscriptions"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2019-08-01/web"
)

// ResourceGroup in which project is deployed
type ResourceGroup struct {
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

// ResourceGroupSource must be implemented to allow chosing resource group
// to which project will be deployed
type ResourceGroupSource interface {
	Select(ctx context.Context, client resources.GroupsClient) (ResourceGroup, error)
}

type resourceGroupSourceOpt struct {
	ResourceGroupSource
}

func (r resourceGroupSourceOpt) withOpt(opts *options) {
	opts.resourceGroupSource = r.ResourceGroupSource
}

// ResourceGroupSourceOpt adds resource group source to context
func ResourceGroupSourceOpt(rq ResourceGroupSource) Option {
	return resourceGroupSourceOpt{rq}
}

// Location represents one of valid Azure regions
type Location string

// LocationSource must be implemented to allow chosing location
// to which project or resource will be deployed
type LocationSource interface {
	Select(ctx context.Context, locations []subscriptions.Location) (Location, error)
}

type locationSourceOpt struct {
	LocationSource
}

func (l locationSourceOpt) withOpt(opts *options) {
	opts.resourceGroupLocationSource = l.LocationSource
}

// ResourceGroupLocationSourceOpt allows setting location source in builder
func ResourceGroupLocationSourceOpt(l LocationSource) Option {
	return locationSourceOpt{l}
}

// StorageAccount represents storage account for project in Azure Cloud.
type StorageAccount string

// StorageAccountSource must be implemented to allow chosing storage account
// for project
type StorageAccountSource interface {
	Select(
		ctx context.Context,
		client storage.AccountsClient,
		context Context,
	) (StorageAccount, error)
}

type storageAccountSourceOpt struct {
	StorageAccountSource
}

func (c storageAccountSourceOpt) withOpt(opts *options) {
	opts.storageAccountSource = c.StorageAccountSource
}

// StorageAccountSourceOpt adds storage account source to builder
func StorageAccountSourceOpt(cs StorageAccountSource) Option {
	return storageAccountSourceOpt{cs}
}

// BlobContainer is a blob store in Azure Cloud
type BlobContainer string

// BlobContainerSource must be implemented to allow chosing storage account
// for project
type BlobContainerSource interface {
	Select(
		ctx context.Context,
		client storage.BlobContainersClient,
		context Context,
	) (BlobContainer, error)
}

type blobContainerSourceOpt struct {
	BlobContainerSource
}

func (c blobContainerSourceOpt) withOpt(opts *options) {
	opts.blobContainerSource = c.BlobContainerSource
}

// BlobContainerSourceOpt adds storage account source to builder
func BlobContainerSourceOpt(cs BlobContainerSource) Option {
	return blobContainerSourceOpt{cs}
}

// SchemaSource must be implemented to provide custom logic for schema
// reading
type SchemaSource interface {
	Select(
		ctx context.Context,
		context Context,
	) (string, error)
}

type schemaSourceOpt struct {
	SchemaSource
}

func (s schemaSourceOpt) withOpt(opts *options) {
	opts.schemaSource = s.SchemaSource
}

// SchemaSourceOpt adds schema source
func SchemaSourceOpt(cs SchemaSource) Option {
	return schemaSourceOpt{cs}
}

// StuccoJSONSource must be implemented to provide custom logic for stucco.json
// reading
type StuccoJSONSource interface {
	Select(
		ctx context.Context,
		context Context,
	) (string, error)
}

type stuccoJSONSourceOpt struct {
	StuccoJSONSource
}

func (s stuccoJSONSourceOpt) withOpt(opts *options) {
	opts.stuccoJSONSource = s.StuccoJSONSource
}

// StuccoJSONSourceOpt adds stucco.json source
func StuccoJSONSourceOpt(cs StuccoJSONSource) Option {
	return stuccoJSONSourceOpt{cs}
}

type schemaOpt string

func (s schemaOpt) withOpt(opts *options) {
	opts.schema = string(s)
}

// SchemaOpt adds schema
func SchemaOpt(s string) Option {
	return schemaOpt(s)
}

type stuccoJSONOpt string

func (s stuccoJSONOpt) withOpt(opts *options) {
	opts.stuccoJSON = string(s)
}

// StuccoJSONOpt adds schema
func StuccoJSONOpt(s string) Option {
	return stuccoJSONOpt(s)
}

// FunctionAppPlan used by function app
type FunctionAppPlan struct {
	Name    string `json:"name"`
	Sku     string `json:"sku"`
	Workers int32  `json:"workers"`
}

// FunctionAppImage represents a docker image with function code
type FunctionAppImage struct {
	Repository string `json:"repository"`
	Registry   string `json:"registry,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"-"`
}

// Validate return an error if image is not valid, otherwise nil
func (i FunctionAppImage) Validate() error {
	if i.Repository == "" {
		return errors.New("image requires repository")
	}
	if i.Registry == "" && i.Username == "" && i.Password == "" {
		return nil
	}
	if i.Username == "" || i.Password == "" {
		return errors.New("both password and username are required for docker registry")
	}
	return nil
}

// FunctionApp used by project
type FunctionApp struct {
	Name     string           `json:"name"`
	Plan     FunctionAppPlan  `json:"plan"`
	Location Location         `json:"location"`
	Image    FunctionAppImage `json:"image"`
	Key      string           `json:"functionKey"`
}

// FunctionAppSource must be implemented to allow chosing
// function app configuration for project
type FunctionAppSource interface {
	Select(
		ctx context.Context,
		context Context,
		client web.BaseClient,
	) (FunctionApp, error)
}

type functionAppOpt struct {
	FunctionAppSource
}

func (f functionAppOpt) withOpt(opts *options) {
	opts.functionAppSource = f.FunctionAppSource
}

// FunctionAppOpt allows overriding source of function configuration source
func FunctionAppOpt(source FunctionAppSource) Option {
	return functionAppOpt{source}
}

// SecretsSource allows setting runtime secrets
type SecretsSource interface {
	Select(
		ctx context.Context,
		context Context,
	) (map[string]string, error)
}

type secretsSourceOpt struct {
	SecretsSource
}

func (s secretsSourceOpt) withOpt(opts *options) {
	opts.secretsSource = s.SecretsSource
}

// SecretsSourceOpt option for builder to set secrets source
func SecretsSourceOpt(source SecretsSource) Option {
	return secretsSourceOpt{source}
}
