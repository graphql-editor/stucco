package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/graphql-editor/stucco/pkg/driver/plugin"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/version"
	"github.com/graphql-go/handler"
	"github.com/rs/cors"
	"k8s.io/apiserver/pkg/server/httplog"
	"k8s.io/klog"
)

const config = "./stucco.json"

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

func loadConfig() (cfg router.Config, err error) {
	st, err := os.Stat(config)
	if err != nil || st.IsDir() {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		if err == nil {
			err = errors.New("./stucco.json is a directory")
		}
		return
	}
	f, err := os.Open(config)
	if err != nil {
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cfg)
	return
}

var versionCheck bool

func main() {
	klog.InitFlags(nil)
	if verb := flag.CommandLine.Lookup("v"); verb != nil {
		l := klog.Level(3)
		verb.DefValue = l.String()
		verbosityLevel := (verb.Value.(*klog.Level))
		*verbosityLevel = l
	}
	flag.BoolVar(&versionCheck, "version", false, "print version and exit")
	flag.Parse()
	if versionCheck {
		fmt.Println(version.Version)
		os.Exit(0)
	}
	cfg, err := loadConfig()
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
	if err := http.ListenAndServe(":8080", nil); err != nil {
		klog.Fatalln(err)
	}
}
