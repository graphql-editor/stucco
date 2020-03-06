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
package cmd

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/graphql-editor/stucco/pkg/driver/plugin"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/graphql-go/handler"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/server/httplog"
	"k8s.io/klog"
)

func withProtocolInContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(
			rw,
			r.WithContext(
				context.WithValue(
					r.Context(),
					router.ProtocolKey, map[string]interface{}{
						"headers": r.Header,
					},
				),
			),
		)
	})
}

func recoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				log.Println(err)
				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte("There was an internal server error"))
			}
		}()
		next.ServeHTTP(rw, r)
	})
}

// startCmd represents the start command
var (
	startConfig string
	startCmd    = &cobra.Command{
		Use:   "start",
		Short: "Start local runner",
		Run: func(cmd *cobra.Command, args []string) {
			var cfg router.Config
			err := utils.LoadConfigFile(startConfig, &cfg)
			if err != nil {
				klog.Fatalln(err)
			}
			cleanupPlugins := plugin.LoadDriverPlugins(plugin.Config{})
			defer cleanupPlugins()
			router, err := router.NewRouter(cfg)
			if err != nil {
				klog.Fatalln(err)
			}
			h := handler.New(&handler.Config{
				Schema:   &router.Schema,
				Pretty:   true,
				GraphiQL: true,
			})
			http.Handle(
				"/graphql",
				recoveryHandler(
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
						}).Handler(
							withProtocolInContext(h),
						),
						httplog.DefaultStacktracePred,
					),
				),
			)
			server := http.Server{
				Addr: ":8080",
			}
			shc := make(chan os.Signal, 1)
			signal.Notify(shc, syscall.SIGTERM)
			go func() {
				<-shc
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				server.Shutdown(ctx)
			}()
			if err := server.ListenAndServe(); err != nil {
				klog.Errorln(err)
			}
		},
	}
)

func init() {
	localCmd.AddCommand(startCmd)
	klogFlagSet := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlagSet)
	if verb := klogFlagSet.Lookup("v"); verb != nil {
		l := klog.Level(3)
		verb.DefValue = l.String()
		verbosityLevel := (verb.Value.(*klog.Level))
		*verbosityLevel = l
	}
	startCmd.Flags().AddGoFlagSet(klogFlagSet)
	startCmd.Flags().StringVarP(&startConfig, "config", "c", "", "path to stucco config")

	localCmd.Flags()
}
