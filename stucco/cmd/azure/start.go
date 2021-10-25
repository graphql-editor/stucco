package azurecmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/spf13/cobra"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/server"
	"github.com/graphql-editor/stucco/pkg/utils"
)

func createBlobLinks(container, account, key, connectionString string) (config, schema string, err error) {
	if connectionString != "" {
		parts := strings.Split(connectionString, ";")
		for _, p := range parts {
			if acc := strings.TrimPrefix(p, "AccountName="); acc != p {
				account = acc
			}
			if k := strings.TrimPrefix(p, "AccountKey="); k != p {
				key = k
			}
		}
	}
	if container == "" || account == "" || key == "" {
		return
	}
	credential, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range []struct {
		fileName string
		dest     *string
	}{
		{"schema.graphql", &schema},
		{"stucco.json", &config},
	} {
		var sasQueryParams azblob.SASQueryParameters
		sasQueryParams, err = azblob.BlobSASSignatureValues{
			Protocol:      azblob.SASProtocolHTTPS,
			ExpiryTime:    time.Now().UTC().Add(48 * time.Hour),
			ContainerName: container,
			BlobName:      d.fileName,
			Permissions:   azblob.BlobSASPermissions{Add: false, Read: true, Write: false}.String(),
		}.NewSASQueryParameters(credential)
		if err != nil {
			return
		}

		qp := sasQueryParams.Encode()
		*d.dest = fmt.Sprintf(
			"https://%s.blob.core.windows.net/%s/%s?%s",
			account,
			container,
			d.fileName,
			qp,
		)
	}
	return
}

type storageDefaults struct {
	account          string
	key              string
	connectionString string
	stuccoFiles      string
}
type azureDefaults struct {
	listenAddress string
	storage       storageDefaults
	schema        string
	config        string
	worker        string
	maxDepth      int
}

var defaults = func() azureDefaults {
	listenPort := "8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenPort = val
	}
	worker := "http://localhost:7071"
	if val, ok := os.LookupEnv("STUCCO_AZURE_WORKER_BASE_URL"); ok {
		worker = val
	}
	maxDepth := int64(0)
	if maxDepthEnv := os.Getenv("STUCCO_AZURE_WORKER_MAX_DEPTH"); maxDepthEnv != "" {
		var err error
		maxDepth, err = strconv.ParseInt(maxDepthEnv, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
	}
	return azureDefaults{
		listenAddress: ":" + listenPort,
		storage: storageDefaults{
			account:          os.Getenv("STUCCO_AZURE_WORKER_STORAGE_ACCOUNT"),
			key:              os.Getenv("STUCCO_AZURE_WORKER_STORAGE_ACCOUNT_KEY"),
			connectionString: os.Getenv("AzureWebJobsStorage"),
			stuccoFiles:      os.Getenv("STUCCO_AZURE_CONTAINER"),
		},
		schema:   os.Getenv("STUCCO_SCHEMA"),
		config:   os.Getenv("STUCCO_CONFIG"),
		worker:   worker,
		maxDepth: int(maxDepth),
	}
}()

// NewStartCommand returns new start command
func NewStartCommand() *cobra.Command {
	var config string
	var schema string
	var worker string
	var listen string
	var saAccount string
	var saKey string
	var saConnectionString string
	var saStuccoFiles string
	var cert string
	var key string
	var maxDepth int
	startCommand := &cobra.Command{
		Use:   "start",
		Short: "Run azure router",
		Run: func(cmd *cobra.Command, args []string) {
			var cfg server.Config
			if config == "" && schema == "" {
				var err error
				config, schema, err = createBlobLinks(saStuccoFiles, saAccount, saKey, saConnectionString)
				if err != nil {
					log.Fatal(err)
				}
			}
			if err := utils.LoadConfigFile(config, &cfg); err != nil {
				log.Fatal(err)
			}
			if schema != "" {
				cfg.Schema = schema
			}
			if maxDepth != 0 {
				cfg.MaxDepth = maxDepth
			}
			var azureAttribs map[string]interface{}
			if cert != "" || key != "" {
				azureAttribs = map[string]interface{}{
					"cert": cert,
					"key":  key,
				}
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
						Type:       server.Azure,
						Attributes: azureAttribs,
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
				Addr:    listen,
			}
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		},
	}
	startCommand.Flags().StringVarP(&config, "config", "c", defaults.config, "Path or url to stucco config")
	startCommand.Flags().StringVarP(&listen, "listen", "l", defaults.listenAddress, "Router listen address")
	startCommand.Flags().StringVarP(&schema, "schema", "s", defaults.schema, "Path or url to stucco schema")
	startCommand.Flags().StringVarP(&worker, "worker", "w", defaults.worker, "Address of azure function worker")
	startCommand.Flags().StringVar(&saAccount, "storage-account", defaults.storage.account, "Storage account to use")
	startCommand.Flags().StringVar(&saKey, "storage-account-key", defaults.storage.key, "Key to storage account")
	startCommand.Flags().StringVar(&saConnectionString, "storage-connection-string", defaults.storage.connectionString, "Connection string to azure web jobs storage")
	startCommand.Flags().StringVar(&saStuccoFiles, "stucco-files-container", defaults.storage.stuccoFiles, "A name of container with stucco files in Azure Storage")
	startCommand.Flags().StringVar(&cert, "cert", "", "Certficate for client cert auth")
	startCommand.Flags().StringVar(&key, "key", "", "Key for client cert auth")
	startCommand.Flags().IntVar(&maxDepth, "max-depth", defaults.maxDepth, "Limit GraphQL recursion")
	return startCommand
}
