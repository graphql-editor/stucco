// Package localcmd is a local command
/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package localcmd

import (
	"flag"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/handlers"
	"github.com/graphql-editor/stucco/pkg/server"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/server/httplog"
	"k8s.io/klog"
)

type klogErrorf struct{}

func (klogErrorf) Errorf(msg string, args ...interface{}) {
	klog.Errorf(msg, args...)
}

// NewStartCommand creates a start command
func NewStartCommand() *cobra.Command {
	var startConfig string
	var schema string
	startCommand := &cobra.Command{
		Use:   "start",
		Short: "Start local runner",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg server.Config
			if err := utils.LoadConfigFile(startConfig, &cfg); err != nil {
				return err
			}
			if schema != "" {
				cfg.Schema = schema
			}
			dri := server.NewDefaultDrivers()
			if err := dri.Load(); err != nil {
				return err
			}
			defer dri.Close()
			h, err := server.New(cfg)
			if err != nil {
				return err
			}
			webhookHandler, err := server.NewWebhookHandler(cfg)
			if err != nil {
				return err
			}
			middleware := func(next http.Handler) http.Handler {
				return handlers.RecoveryHandler(
					httplog.WithLogging(
						cors.New(cors.Options{
							AllowedOrigins: []string{"*"},
							AllowedMethods: []string{
								http.MethodHead,
								http.MethodGet,
								http.MethodPost,
								http.MethodPut,
								http.MethodPatch,
								http.MethodDelete,
							},
							AllowedHeaders:   []string{"*"},
							AllowCredentials: true,
						}).Handler(next),
						httplog.DefaultStacktracePred,
					),
					klogErrorf{},
				)
			}
			h = middleware(h)
			webhookHandler = middleware(webhookHandler)
			srv := server.Server{
				Handler:        h,
				WebhookHandler: webhookHandler,
				Addr:           ":8080",
			}
			return srv.ListenAndServe()
		},
	}
	klogFlagSet := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlagSet)
	if verb := klogFlagSet.Lookup("v"); verb != nil {
		l := klog.Level(3)
		verb.DefValue = l.String()
		verbosityLevel := (verb.Value.(*klog.Level))
		*verbosityLevel = l
	}
	startCommand.Flags().AddGoFlagSet(klogFlagSet)
	startCommand.Flags().StringVarP(&startConfig, "config", "c", "", "path to stucco config")
	startCommand.Flags().StringVarP(&schema, "schema", "s", "", "path to stucco config")
	return startCommand
}
