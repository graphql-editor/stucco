package azurecmd

import (
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/server"
	"github.com/graphql-editor/stucco/pkg/utils"
)

type storageDefaults struct {
	account          string
	key              string
	connectionString string
}
type azureDefaults struct {
	listenAddress string
	storage       storageDefaults
	schema        string
	config        string
	worker        string
}

var defaults = func() azureDefaults {
	listenPort := "8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenPort = val
	}
	schema := "schema.graphql"
	if val, ok := os.LookupEnv("STUCCO_SCHEMA"); ok {
		schema = val
	}
	config := "stucco.json"
	if val, ok := os.LookupEnv("STUCCO_CONFIG"); ok {
		config = val
	}
	worker := "http://localhost:7071"
	if val, ok := os.LookupEnv("STUCCO_AZURE_WORKER_BASE_URL"); ok {
		worker = val
	}
	return azureDefaults{
		listenAddress: ":" + listenPort,
		storage: storageDefaults{
			account:          os.Getenv("STUCCO_AZURE_WORKER_STORAGE_ACCOUNT"),
			key:              os.Getenv("STUCCO_AZURE_WORKER_STORAGE_ACCOUNT_KEY"),
			connectionString: os.Getenv("AzureWebJobsStorage"),
		},
		schema: schema,
		config: config,
		worker: worker,
	}
}()

// NewStartCommand returns new start command
func NewStartCommand() *cobra.Command {
	var config string
	var schema string
	var worker string
	var listen string
	var key string
	var saAccount string
	var saKey string
	var saConnectionString string
	startCommand := &cobra.Command{
		Use:   "start",
		Short: "Run azure router",
		Run: func(cmd *cobra.Command, args []string) {
			var cfg server.Config
			if err := utils.LoadConfigFile(config, &cfg); err != nil {
				log.Fatal(err)
			}
			if schema != "" {
				cfg.Schema = schema
			}
			h, err := server.New(server.Config{
				Config: cfg.Config,
				Drivers: server.Drivers{
					{Type: server.Plugin},
					{
						Config: driver.Config{
							Provider: "azure",
							Runtime:  "function",
						},
						Type: server.Azure,
						Attributes: map[string]interface{}{
							"worker":      worker,
							"functionKey": key,
							"storage": map[string]interface{}{
								"account":          saAccount,
								"key":              saKey,
								"connectionString": saConnectionString,
							},
						},
					},
				},
				DefaultEnvironment: router.Environment{
					Provider: "azure",
					Runtime:  "function",
				},
			})
			if err != nil {
				log.Fatal(err)
			}
			srv := server.Server{
				Handler: h,
				Addr:    defaults.listenAddress,
			}
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		},
	}
	startCommand.Flags().StringVarP(&config, "config", "c", defaults.config, "Path or url to stucco config")
	startCommand.Flags().StringVarP(&key, "function-key", "k", "", "If function is deployed with function authLevel function key is required")
	startCommand.Flags().StringVarP(&listen, "listen", "l", defaults.listenAddress, "Router listen address")
	startCommand.Flags().StringVarP(&schema, "schema", "s", defaults.schema, "Path or url to stucco schema")
	startCommand.Flags().StringVarP(&worker, "worker", "w", defaults.worker, "Address of azure function worker")
	startCommand.Flags().StringVar(&saAccount, "storage-account", defaults.storage.account, "Storage account to use")
	startCommand.Flags().StringVar(&saKey, "storage-account-key", defaults.storage.key, "Key to storage account")
	startCommand.Flags().StringVar(&saConnectionString, "storage-connection-string", defaults.storage.connectionString, "Connection string to azure web jobs storage")
	return startCommand
}
