package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/graphql-editor/stucco/pkg/driver/plugin"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/version"
	"github.com/graphql-go/handler"
	"github.com/rs/cors"
	"k8s.io/apiserver/pkg/server/httplog"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"
)

const config = "./stucco"

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

type decodeFunc func([]byte, interface{}) error

func yamlUnmarshal(b []byte, v interface{}) error {
	return yaml.Unmarshal(b, v)
}

var supportedExtension = map[string]decodeFunc{
	".json": json.Unmarshal,
	".yaml": yamlUnmarshal,
	".yml":  yamlUnmarshal,
}

func loadConfig() (cfg router.Config, err error) {
	var decode decodeFunc
	var st os.FileInfo
	var configPath string
	for k, v := range supportedExtension {
		st, err = os.Stat(config + k)
		if err == nil {
			configPath = config + k
			decode = v
		}
		if !os.IsNotExist(err) {
			break
		}
	}
	if err != nil || st.IsDir() {
		if os.IsNotExist(err) {
			err = fmt.Errorf("could not find stucco config in current directory")
			return
		}
		if err == nil {
			err = fmt.Errorf("%s is a directory", configPath)
		}
		return
	}
	b, err := ioutil.ReadFile(configPath)
	if err == nil {
		err = decode(b, &cfg)
	}
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
}
