package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/plugin"
	"github.com/graphql-editor/stucco/pkg/handlers"
	gqlhandler "github.com/graphql-editor/stucco/pkg/handlers"
	azuredriver "github.com/graphql-editor/stucco/pkg/providers/azure/driver"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/security"
	"k8s.io/klog"
)

// DriverKind represents one of implemented drivers for handling functions
type DriverKind uint8

// Enums representing supported drivers
const (
	Unknown DriverKind = iota
	Plugin
	Azure
)

// UnmarshalJSON implements Unmarshaler
func (d *DriverKind) UnmarshalJSON(b []byte) (err error) {
	*d = Unknown
	var s string
	if err = json.Unmarshal(b, &s); err == nil {
		switch s {
		case "plugin":
			*d = Plugin
		case "azure":
			*d = Azure
		default:
			err = errors.New("invalid DriverKind")
		}
	}
	return
}

// MarshalJSON implements json.Marshaler
func (d DriverKind) MarshalJSON() (b []byte, err error) {
	var s string
	switch d {
	case Plugin:
		s = "plugin"
	case Azure:
		s = "azure"
	default:
		return nil, errors.New("invalid DriverKind")
	}
	return json.Marshal(s)
}

type staticKey string

func (k staticKey) GetKey(function string) (string, error) {
	return string(k), nil
}

// Driver is a driver definition for server
type Driver struct {
	driver.Config
	Type       DriverKind             `json:"type"`
	Attributes map[string]interface{} `json:"-"`
	Optional   bool                   `json:"optional"`
	closer     io.Closer
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Driver) UnmarshalJSON(b []byte) error {
	type driver Driver
	dd := (*driver)(d)
	err := json.Unmarshal(b, dd)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &d.Attributes)
}

// Close implements io.Closer
func (d *Driver) Close() (err error) {
	if d.closer != nil {
		err = d.closer.Close()
	}
	return
}

func (d *Driver) pluginLoad() error {
	cleanup := plugin.LoadDriverPlugins(plugin.Config{})
	d.closer = closeNoErrFn(cleanup)
	return nil
}

type azureClient struct {
	rt http.RoundTripper
	azuredriver.ProtobufClient
}

// New returns new driver using protobuf protocol
func (a azureClient) New(u, f string) driver.Driver {
	if a.HTTPClient == nil && a.rt != nil {
		a.HTTPClient = &http.Client{
			Transport: a.rt,
		}
	}
	return a.ProtobufClient.New(u, f)
}

func (d *Driver) azureLoad() error {
	var worker string
	var cert string
	var key string
	set := func(dst *string, k string, m map[string]interface{}) {
		if m == nil {
			return
		}
		if val, ok := m[k].(string); ok {
			*dst = val
		}
	}
	set(&worker, "worker", d.Attributes)
	set(&cert, "cert", d.Attributes)
	set(&key, "key", d.Attributes)
	cli := azureClient{}
	dri := &azuredriver.Driver{
		BaseURL:      worker,
		WorkerClient: &cli,
	}
	if cert != "" && key != "" {
		auth := security.Auth{
			Cert: &security.CertAuth{
				Cert:         []byte(cert),
				Key:          []byte(key),
				Renegotation: tls.RenegotiateOnceAsClient,
			},
		}
		rt, err := auth.RoundTripper()
		if err != nil {
			return err
		}
		cli.rt = rt
	}
	driver.Register(d.Config, dri)
	return nil
}

// Load loads a known driver type with config
func (d *Driver) Load() error {
	switch d.Type {
	case Plugin:
		return d.pluginLoad()
	case Azure:
		return d.azureLoad()
	}
	return errors.New("unsupported DriverKind")
}

// DriversCloseError is a list of errors
type DriversCloseError []error

func (d DriversCloseError) Error() string {
	errMsgs := make([]string, 0, len(d))
	for _, e := range d {
		errMsgs = append(errMsgs, e.Error())
	}
	return "error closing drivers: " + strings.Join(errMsgs, ", ")
}

// Drivers is a list of supported by router
type Drivers []Driver

// Load loads known drivers with their configuration
func (d *Drivers) Load() error {
	for i := range *d {
		if err := (*d)[i].pluginLoad(); err != nil && !(*d)[i].Optional {
			return err
		}
	}
	return nil
}

// Close implements io.Closer
func (d *Drivers) Close() (err error) {
	var errs DriversCloseError
	for _, dr := range *d {
		if err := dr.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// NewDefaultDrivers returns default drivers which include local and azure driver for localhost
func NewDefaultDrivers() Drivers {
	return Drivers{
		{Type: Plugin},
		{
			Config: driver.Config{
				Provider: "azure",
				Runtime:  "function",
			},
			Type: Azure,
			Attributes: map[string]interface{}{
				"worker": "http://localhost",
			},
		},
	}
}

type closeNoErrFn func()

func (c closeNoErrFn) Close() error {
	c()
	return nil
}

func checkPointerBoolDefaultTrue(b *bool) bool {
	if b == nil {
		return true
	}
	return *b
}

// Config is a GraphQL http server configuration
type Config struct {
	router.Config
	Pretty             *bool              `json:"pretty"`
	GraphiQL           *bool              `json:"graphiql"`
	DefaultEnvironment router.Environment `json:"defaultEnvironment"`
}

// UnmarshalJSON implements json unmarshaler
func (c *Config) UnmarshalJSON(b []byte) error {
	type config Config
	if err := json.Unmarshal(b, &c.Config); err != nil {
		return err
	}
	return json.Unmarshal(b, (*config)(c))
}

// UnmarshalYAML implements yaml unmarshaler
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type config Config
	if err := unmarshal(&c.Config); err != nil {
		return err
	}
	return unmarshal((*config)(c))
}

// New returns new handler for graphql server
func New(c Config) (httpHandler http.Handler, err error) {
	if err == nil {
		if c.DefaultEnvironment.Provider != "" || c.DefaultEnvironment.Runtime != "" {
			router.SetDefaultEnvironment(c.DefaultEnvironment)
		}
	}
	var rt router.Router
	if err == nil {
		rt, err = router.NewRouter(c.Config)
	}
	if err == nil {
		httpHandler = handlers.WithProtocolInContext(gqlhandler.New(gqlhandler.Config{
			Schema:   &rt.Schema,
			Pretty:   checkPointerBoolDefaultTrue(c.Pretty),
			GraphiQL: checkPointerBoolDefaultTrue(c.GraphiQL),
		}))
	}
	return
}

// Server default simple server that has two endpoints. /graphql which uses Handler as a handler
// and /health that uses Health as a handler or just returns 200.
// It handles SIGTERM.
type Server struct {
	Handler http.Handler
	Health  http.Handler
	Addr    string
}

func (s *Server) health(rw http.ResponseWriter, req *http.Request) {
	if s.Health != nil {
		s.Health.ServeHTTP(rw, req)
		return
	}
	fmt.Fprint(rw, "OK")
}

// ListenAndServe is a simple wrapper around http.Server.ListenAndServe with two endpoints. It is blocking
func (s *Server) ListenAndServe() error {
	mux := http.ServeMux{}
	mux.Handle("/graphql", s.Handler)
	mux.HandleFunc("/health", s.health)
	srv := http.Server{
		Addr:    s.Addr,
		Handler: &mux,
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			klog.Error(err)
		}
	}()
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}
