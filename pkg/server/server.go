package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/plugin"
	"github.com/graphql-editor/stucco/pkg/handlers"
	azuredriver "github.com/graphql-editor/stucco/pkg/providers/azure/driver"
	"github.com/graphql-editor/stucco/pkg/router"
	gqlhandler "github.com/graphql-go/handler"
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

type staticKey string

func (k staticKey) GetKey(function string) (string, error) {
	return string(k), nil
}

// Driver is a driver definition for server
type Driver struct {
	driver.Config
	Type       DriverKind             `json:"type"`
	Attributes map[string]interface{} `json:"-"`
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
func (d Driver) Close() (err error) {
	if d.closer != nil {
		err = d.closer.Close()
	}
	return
}

func (d Driver) pluginLoad() (Driver, error) {
	cleanup := plugin.LoadDriverPlugins(plugin.Config{})
	d.closer = closeNoErrFn(cleanup)
	return d, nil
}

func (d Driver) azureLoad() (Driver, error) {
	var worker string
	var key string
	var saAccount string
	var saKey string
	var saConnectionString string
	set := func(dst *string, k string, m map[string]interface{}) {
		if m == nil {
			return
		}
		if val, ok := m[k].(string); ok {
			*dst = val
		}
	}
	set(&worker, "worker", d.Attributes)
	set(&key, "functionKey", d.Attributes)
	if v, ok := d.Attributes["storage"].(map[string]interface{}); ok {
		set(&saAccount, "account", v)
		set(&saKey, "key", v)
		set(&saConnectionString, "connectionString", v)
	}
	dri := &azuredriver.Driver{
		BaseURL: worker,
	}
	var kr azuredriver.KeyReader
	if saAccount != "" && saKey != "" || saConnectionString != "" {
		kr = &azuredriver.StorageHostKeyReader{
			Driver:           dri,
			Account:          saAccount,
			Key:              saKey,
			ConnectionString: saConnectionString,
		}
	}
	if kr == nil && key != "" {
		kr = staticKey(key)
	}
	dri.WorkerClient = &azuredriver.ProtobufClient{
		KeyReader: kr,
	}
	driver.Register(d.Config, dri)
	return d, nil
}

// Load loads a known driver type with config
func (d Driver) Load() (Driver, error) {
	switch d.Type {
	case Plugin:
		return d.pluginLoad()
	case Azure:
		return d.azureLoad()
	}
	return Driver{}, errors.New("unsupported DriverKind")
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

func (d *Drivers) addDriver(dri Driver, err error) error {
	if err == nil {
		*d = append(*d, dri)
	}
	return err
}

// Load loads known drivers with their configuration
func (d *Drivers) Load() (io.Closer, error) {
	newDrivers := make(Drivers, 0, len(*d))
	for _, dri := range *d {
		var err error
		switch dri.Type {
		case Plugin:
			err = d.addDriver(dri.pluginLoad())
		case Azure:
			err = d.addDriver(dri.azureLoad())
		default:
			err = errors.New("unsupported DriverKind")
		}
		if err != nil {
			defer newDrivers.Close()
			return nil, err
		}
	}
	return newDrivers, nil
}

// Close implements io.Closer
func (d Drivers) Close() (err error) {
	var errs DriversCloseError
	for _, dr := range d {
		if err := dr.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

var defaultDrivers = Drivers{
	{Type: Plugin},
	{
		Config: driver.Config{
			Provider: "azure",
			Runtime:  "function",
		},
		Type: Azure,
	},
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
	Drivers            Drivers            `json:"drivers"`
	DefaultEnvironment router.Environment `json:"defaultEnvironment"`
}

// New returns new handler for graphql server
func New(c Config) (httpHandler http.Handler, err error) {
	drivers := c.Drivers
	if len(drivers) == 0 {
		drivers = defaultDrivers
	}
	driversCleanup, err := drivers.Load()
	if err == nil {
		defer driversCleanup.Close()
		if c.DefaultEnvironment.Provider != "" || c.DefaultEnvironment.Runtime != "" {
			router.SetDefaultEnvironment(c.DefaultEnvironment)
		}
	}
	var rt router.Router
	if err == nil {
		rt, err = router.NewRouter(c.Config)
	}
	if err == nil {
		httpHandler = handlers.WithProtocolInContext(gqlhandler.New(&gqlhandler.Config{
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
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
