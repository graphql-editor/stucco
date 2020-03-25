package httptrigger

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/graphql-editor/azure-functions-golang-worker/api"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/handlers"
	azuredriver "github.com/graphql-editor/stucco/pkg/providers/azure/driver"
	azurehandler "github.com/graphql-editor/stucco/pkg/providers/azure/handler"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/utils"
	gqlhandler "github.com/graphql-go/handler"
)

var (
	lock    sync.RWMutex
	config  string
	handler azurehandler.Handler
)

// HTTPTrigger is an example httpTrigger
type HTTPTrigger struct {
	Request  *api.Request `azfunc:"httpTrigger"`
	Response api.Response `azfunc:"res"`
}

// Run implements function behaviour
func (h *HTTPTrigger) Run(ctx context.Context, logger api.Logger) {
	handler, err := getHandler()
	if err != nil {
		logger.Errorf("could not get handler: %v", err)
		h.Response = api.Response{
			Headers: http.Header{
				"content-type": []string{"text/plain"},
			},
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}
		return
	}
	h.Response = handler.ServeHTTP(ctx, logger, h.Request)
}

func configValue() string {
	stuccoConfig := ""
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "STUCCO_") {
			stuccoConfig += v + ";"
		}
	}
	return stuccoConfig
}

// Config for GraphQL router using azure worker functions
type Config struct {
	Account, Key, ConnectionString, StuccoConfig, Schema string
}

// NewHandler creates new http handler using azure worker functions
func NewHandler(c Config) (httpHandler http.Handler, err error) {
	driver.Register(driver.Config{
		Provider: "azure",
		Runtime:  "function",
	}, &azuredriver.Driver{
		WorkerClient: &azuredriver.ProtobufClient{
			KeyReader: &azuredriver.StorageHostKeyReader{
				Account:          c.Account,
				Key:              c.Key,
				ConnectionString: c.ConnectionString,
			},
		},
	})
	router.SetDefaultEnvironment(router.Environment{
		Provider: "azure",
		Runtime:  "function",
	})
	var cfg router.Config
	if err = utils.LoadConfigFile(c.StuccoConfig, &cfg); err == nil && c.Schema != "" {
		cfg.Schema = c.Schema
	}
	var rt router.Router
	if err == nil {
		rt, err = router.NewRouter(cfg)
	}
	if err == nil {
		httpHandler = handlers.WithProtocolInContext(gqlhandler.New(&gqlhandler.Config{
			Schema:   &rt.Schema,
			Pretty:   true,
			GraphiQL: true,
		}))
	}
	return
}

// getHandler is used by http trigger for router running as a function
// in azure functions host
func getHandler() (rhandler azurehandler.Handler, err error) {
	lock.RLock()
	rhandler = handler
	currentConfig := config
	lock.RUnlock()
	cv := configValue()
	if rhandler.Handler == nil || currentConfig != cv {
		var httpHandler http.Handler
		httpHandler, err = NewHandler(Config{
			Account:          os.Getenv("STUCCO_AZURE_WORKER_STORAGE_ACCOUNT"),
			Key:              os.Getenv("STUCCO_AZURE_WORKER_STORAGE_ACCOUNT_KEY"),
			ConnectionString: os.Getenv("AzureWebJobsStorage"),
		})
		if err == nil {
			rhandler = azurehandler.Handler{
				Handler: httpHandler,
			}
			lock.Lock()
			handler = rhandler
			config = cv
			lock.Unlock()
		}
	}
	return
}
