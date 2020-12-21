package azurecmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/graphql-editor/stucco/pkg/providers/azure/function/graphql/httptrigger"
)

// NewStartCommand returns new start command
func NewStartCommand() *cobra.Command {
	var config string
	var schema string
	var worker string
	var listen string
	var key string
	var saAccount string
	var saKey string
	startCommand := &cobra.Command{
		Use:   "start",
		Short: "Run azure router locally",
		Run: func(cmd *cobra.Command, args []string) {
			if key != "" {
				os.Setenv("STUCCO_AZURE_WORKER_KEY", key)
			}
			os.Setenv("STUCCO_AZURE_WORKER_BASE_URL", worker)
			handler, err := httptrigger.NewHandler(httptrigger.Config{
				StuccoConfig: config,
				Schema:       schema,
				Account:      saAccount,
				Key:          saKey,
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			http.Handle("/graphql", handler)
			srv := http.Server{
				Addr: listen,
			}
			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt)
			go func() {
				<-c
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if err := srv.Shutdown(ctx); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}()
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	startCommand.Flags().StringVarP(&config, "config", "c", "stucco.json", "Path or url to stucco config")
	startCommand.Flags().StringVarP(&key, "function-key", "k", "", "If function is deployed with function authLevel function key is required")
	startCommand.Flags().StringVarP(&listen, "listen", "l", ":8080", "Router listen address")
	startCommand.Flags().StringVarP(&schema, "schema", "s", "schema.graphql", "Path or url to stucco schema")
	startCommand.Flags().StringVarP(&worker, "worker", "w", "http://localhost:8081", "Address of azure function worker")
	startCommand.Flags().StringVar(&saAccount, "storage-account", "", "Storage account to use")
	startCommand.Flags().StringVar(&saKey, "storage-account-key", "", "Key to storage account")
	return startCommand
}
