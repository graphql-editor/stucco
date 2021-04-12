package driver

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/protohttp"
	"github.com/graphql-editor/stucco/pkg/types"
)

// WorkerClient creates new protobuf for communication with workers
type WorkerClient interface {
	New(url, fname string) driver.Driver
}

// HTTPClient used by azure client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ProtobufClient is a worker client using protobuf protocol
type ProtobufClient struct {
	HTTPClient
	FunctionName string
}

// Post implemention for azure worker protobuf communication
func (p ProtobufClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err == nil {
		req.Header.Add("content-type", contentType)
		client := p.HTTPClient
		if client == nil {
			client = http.DefaultClient
		}
		resp, err = client.Do(req)
	}
	return
}

// New returns new driver using protobuf protocol
func (p ProtobufClient) New(u, f string) driver.Driver {
	p.FunctionName = f
	return &protohttp.Client{
		HTTPClient: p,
		URL:        u,
	}
}

// Driver implements stucco driver interface calling protobuf workers over http
// with configurable workers base url
type Driver struct {
	BaseURL     string
	FunctionURL map[string]string
	WorkerClient
}

// EndpointName create endpoint name from string
func EndpointName(p string) string {
	return p
}

func normalizeFuncName(fn string) string {
	fn = strings.ReplaceAll(fn, ".", "_")
	fn = strings.ReplaceAll(fn, "/", "_")
	fn = strings.ToUpper(fn)
	return fn
}

func (d *Driver) envFuncURL(fName string) (*url.URL, error) {
	envURL := d.BaseURL
	if envURL == "" {
		if val, ok := os.LookupEnv("STUCCO_AZURE_WORKER_BASE_URL"); ok {
			envURL = val
		}
	}
	normalizedFuncName := normalizeFuncName(fName)
	if funcURL := os.Getenv("STUCCO_AZURE_" + normalizedFuncName + "_URL"); funcURL != "" {
		envURL = funcURL
	}
	if d.FunctionURL != nil {
		if val, ok := d.FunctionURL[fName]; ok {
			envURL = val
		}
	}
	if envURL == "" {
		return nil, errors.New("base url for function not set")
	}
	return url.Parse(envURL)
}

func (d *Driver) newClient(url, fName string) driver.Driver {
	workerClient := d.WorkerClient
	if workerClient == nil {
		workerClient = &ProtobufClient{}
	}
	return workerClient.New(url, fName)
}

func (d *Driver) baseURL(f types.Function) (us string, err error) {
	u, err := d.envFuncURL(f.Name)
	if err == nil {
		u.Path = EndpointName(f.Name)
		us = u.String()
	}
	return
}

// SetSecrets implements driver.Driver
func (d *Driver) SetSecrets(in driver.SetSecretsInput) driver.SetSecretsOutput {
	// noop, secrets must be sed during deployment
	return driver.SetSecretsOutput{}
}

func (d *Driver) functionClient(f types.Function) (client driver.Driver, derr *driver.Error) {
	url, err := d.baseURL(f)
	if err != nil {
		derr = &driver.Error{
			Message: err.Error(),
		}
		return
	}
	client = d.newClient(url, f.Name)
	return
}

// FieldResolve implements driver.Driver
func (d *Driver) FieldResolve(in driver.FieldResolveInput) driver.FieldResolveOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.FieldResolveOutput{
			Error: err,
		}
	}
	return client.FieldResolve(in)
}

// InterfaceResolveType implements driver.Driver
func (d *Driver) InterfaceResolveType(in driver.InterfaceResolveTypeInput) driver.InterfaceResolveTypeOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.InterfaceResolveTypeOutput{
			Error: err,
		}
	}
	return client.InterfaceResolveType(in)
}

// ScalarParse implements driver.Driver
func (d *Driver) ScalarParse(in driver.ScalarParseInput) driver.ScalarParseOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.ScalarParseOutput{
			Error: err,
		}
	}
	return client.ScalarParse(in)
}

// ScalarSerialize implements driver.Driver
func (d *Driver) ScalarSerialize(in driver.ScalarSerializeInput) driver.ScalarSerializeOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.ScalarSerializeOutput{
			Error: err,
		}
	}
	return client.ScalarSerialize(in)
}

// UnionResolveType implements driver.Driver
func (d *Driver) UnionResolveType(in driver.UnionResolveTypeInput) driver.UnionResolveTypeOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.UnionResolveTypeOutput{
			Error: err,
		}
	}
	return client.UnionResolveType(in)
}

// Stream implements driver.Driver
func (d *Driver) Stream(in driver.StreamInput) driver.StreamOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.StreamOutput{
			Error: err,
		}
	}
	return client.Stream(in)
}

// SubscriptionConnection implements driver.Driver
func (d *Driver) SubscriptionConnection(in driver.SubscriptionConnectionInput) driver.SubscriptionConnectionOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.SubscriptionConnectionOutput{
			Error: err,
		}
	}
	return client.SubscriptionConnection(in)
}

// SubscriptionListen implements driver.Driver
func (d *Driver) SubscriptionListen(in driver.SubscriptionListenInput) driver.SubscriptionListenOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.SubscriptionListenOutput{
			Error: err,
		}
	}
	return client.SubscriptionListen(in)
}
