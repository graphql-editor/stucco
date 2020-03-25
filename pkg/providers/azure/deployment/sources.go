package deployment

/*
 All default queriers are just here for headless runs. Meaing they return
 data that was already provided through environment/command line.
 They must fail otherwise without any changes being made.
*/

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-06-01/subscriptions"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2019-08-01/web"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/pkg/errors"
)

type defaultResourceGroupSource struct {
	opts *options
}

func (d defaultResourceGroupSource) Select(ctx context.Context, groupsClient resources.GroupsClient) (rg ResourceGroup, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	if d.opts == nil || d.opts.resourceGroup == "" {
		err = errors.Errorf("could not find a resource group name")
		return
	}
	rg = ResourceGroup{
		Name: d.opts.resourceGroup,
	}
	return
}

type defaultResourceGroupLocationSource struct {
	opts *options
}

func (d defaultResourceGroupLocationSource) Select(ctx context.Context, locations []subscriptions.Location) (location Location, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	if d.opts == nil || d.opts.resourceGroupLocation == "" {
		err = errors.Errorf("location not set")
		return
	}
	location = Location(d.opts.resourceGroupLocation)
	return
}

type defaultStorageAccountSource struct {
	opts *options
}

func (d defaultStorageAccountSource) Select(
	ctx context.Context,
	client storage.AccountsClient,
	context Context,
) (a StorageAccount, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	if d.opts == nil || d.opts.storageAccount == "" {
		err = errors.Errorf("storage account name missing")
		return
	}
	a = StorageAccount(d.opts.storageAccount)
	return
}

type defaultBlobContainerSource struct {
	opts *options
}

func (d defaultBlobContainerSource) Select(
	ctx context.Context,
	client storage.BlobContainersClient,
	context Context,
) (b BlobContainer, err error) {
	if err = ctx.Err(); err != nil {
		return
	}

	if d.opts == nil || d.opts.blobContainer == "" {
		err = errors.Errorf("blob container name missing")
		return
	}
	b = BlobContainer(d.opts.blobContainer)
	return
}

func stringFromURL(url *url.URL) (s string, err error) {
	var r io.ReadCloser
	if url.Scheme == "https" || url.Scheme == "http" {
		var resp *http.Response
		resp, err = http.Get(url.String())
		r = resp.Body
	} else {
		r, err = os.Open(url.String())
	}
	if err == nil {
		defer func() {
			ferr := r.Close()
			if err == nil {
				err = ferr
			}
		}()
		var b []byte
		b, err = ioutil.ReadAll(r)
		if err == nil {
			s = string(b)
		}
	}
	return
}

type defaultSchemaSource struct {
	opts *options
}

func (d defaultSchemaSource) Select(
	ctx context.Context,
	context Context,
) (schema string, err error) {
	// if schema is empty in opts do nothing for now
	// as it is possible that we do not want to upload new schema
	if d.opts == nil || d.opts.schema == "" {
		return
	}
	u, err := url.Parse(d.opts.schema)
	if err != nil {
		_, perr := parser.Parse(parser.ParseParams{Source: d.opts.schema})
		if perr == nil {
			schema = d.opts.schema
			return
		}
		err = errors.Wrap(
			errors.Wrap(perr, err.Error()),
			"schema is not a valid url or graphql schema",
		)
	} else {
		schema, err = stringFromURL(u)
	}
	return
}

type defaultStuccoJSONSource struct {
	opts *options
}

func (d defaultStuccoJSONSource) Select(
	ctx context.Context,
	context Context,
) (stuccoJSON string, err error) {
	// if stucco.json is empty in opts do nothing for now
	// as it is possible that we do not want to upload new config
	if d.opts == nil || d.opts.stuccoJSON == "" {
		return
	}
	u, err := url.Parse(d.opts.stuccoJSON)
	if err != nil {
		var dummy dummy
		jerr := json.Unmarshal([]byte(d.opts.stuccoJSON), &dummy)
		if jerr == nil {
			stuccoJSON = d.opts.stuccoJSON
			return
		}
		err = errors.Wrap(
			errors.Wrap(jerr, err.Error()),
			"stucco.json is not a valid url or json",
		)
	} else {
		stuccoJSON, err = stringFromURL(u)
	}
	return
}

// dummy just to test if json parses without any real allocations
type dummy struct{}

func (d *dummy) UnmarshalJSON([]byte) error {
	return nil
}

type defaultFunctionAppSource struct {
	opts *options
}

func (d defaultFunctionAppSource) Select(
	ctx context.Context,
	context Context,
	client web.BaseClient,
) (fp FunctionApp, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	if d.opts == nil {
		err = errors.Errorf("could not find function app config")
		return
	}
	if d.opts.functionAppName == "" {
		err = errors.Errorf("could not find a function app name")
		return
	}
	if d.opts.functionAppPlanName == "" {
		err = errors.Errorf("could not find a function app plan name")
		return
	}
	if d.opts.functionAppPlanSku == "" {
		err = errors.Errorf("could not find a function app plan sku")
		return
	}
	if d.opts.functionAppPlanWorkers == 0 {
		err = errors.Errorf("could not find a function app plan workers count")
		return
	}
	if err = d.opts.functionAppImage.Validate(); err != nil {
		return
	}
	fp = FunctionApp{
		Name: d.opts.functionAppName,
		Plan: FunctionAppPlan{
			Name:    d.opts.functionAppPlanName,
			Sku:     d.opts.functionAppPlanSku,
			Workers: d.opts.functionAppPlanWorkers,
		},
		Location: Location(d.opts.functionAppLocation),
		Image:    d.opts.functionAppImage,
	}
	return
}

type defaultSecretsSource struct {
	opts *options
}

func (d defaultSecretsSource) Select(
	ctx context.Context,
	context Context,
) (secrets map[string]string, err error) {
	return
}
