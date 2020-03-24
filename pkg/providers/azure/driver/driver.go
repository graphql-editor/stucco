package driver

import (
	"crypto/sha256"
	"encoding/hex"
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
	New(url string) driver.Driver
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
	var authCode string
	if err == nil {
		authCode = os.Getenv("STUCCO_AZURE_WORKER_KEY")
		if funcCode := os.Getenv("STUCCO_AZURE_" + normalizeFuncName(p.FunctionName) + "_KEY"); funcCode != "" {
			authCode = funcCode
		}
	}
	if err == nil {
		if authCode != "" {
			req.Header.Add("X-Functions-Key", authCode)
		}
		resp, err = p.Do(req)
	}
	return
}

// New returns new driver using protobuf protocol
func (p ProtobufClient) New(u string) driver.Driver {
	return &protohttp.Client{
		HTTPClient: p,
		URL:        u,
	}
}

// Driver implements stucco driver interface calling protobuf workers over http
// with configurable workers base url
type Driver struct {
	WorkerClient
}

// EndpointName create endpoint name from string
func EndpointName(p string) string {
	b := sha256.Sum256([]byte(p))
	return hex.EncodeToString(b[:])
}

func normalizeFuncName(fn string) string {
	fn = strings.ReplaceAll(fn, ".", "_")
	fn = strings.ReplaceAll(fn, "/", "_")
	fn = strings.ToUpper(fn)
	return fn
}

func (d *Driver) newClient(url, fName string) driver.Driver {
	workerClient := d.WorkerClient
	if workerClient == nil {
		workerClient = &ProtobufClient{
			FunctionName: fName,
		}
	}
	return workerClient.New(url)
}

func (d *Driver) baseURL(f types.Function) (us string, err error) {
	envURL := os.Getenv("STUCCO_AZURE_WORKER_BASE_URL")
	if funcURL := os.Getenv("STUCCO_AZURE_" + normalizeFuncName(f.Name) + "_URL"); funcURL != "" {
		envURL = funcURL
	}
	u, err := url.Parse(envURL)
	if err == nil {
		us = u.String()
	}
	return
}

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

func (d *Driver) FieldResolve(in driver.FieldResolveInput) driver.FieldResolveOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.FieldResolveOutput{
			Error: err,
		}
	}
	return client.FieldResolve(in)
}

func (d *Driver) InterfaceResolveType(in driver.InterfaceResolveTypeInput) driver.InterfaceResolveTypeOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.InterfaceResolveTypeOutput{
			Error: err,
		}
	}
	return client.InterfaceResolveType(in)
}

func (d *Driver) ScalarParse(in driver.ScalarParseInput) driver.ScalarParseOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.ScalarParseOutput{
			Error: err,
		}
	}
	return client.ScalarParse(in)
}
func (d *Driver) ScalarSerialize(in driver.ScalarSerializeInput) driver.ScalarSerializeOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.ScalarSerializeOutput{
			Error: err,
		}
	}
	return client.ScalarSerialize(in)
}
func (d *Driver) UnionResolveType(in driver.UnionResolveTypeInput) driver.UnionResolveTypeOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.UnionResolveTypeOutput{
			Error: err,
		}
	}
	return client.UnionResolveType(in)
}
func (d *Driver) Stream(in driver.StreamInput) driver.StreamOutput {
	client, err := d.functionClient(in.Function)
	if err != nil {
		return driver.StreamOutput{
			Error: err,
		}
	}
	return client.Stream(in)
}
