package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog"

	"github.com/gorilla/websocket"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

// ResultCallbackFn called with result
type ResultCallbackFn func(ctx context.Context, params *graphql.Params, result *graphql.Result, responseBody []byte)

// Config new handler
type Config struct {
	Schema   *graphql.Schema
	Pretty   bool
	GraphiQL bool
}

// subscriptionHandler is a websocket handler
type subscriptionHandler struct {
	pretty bool
	schema *graphql.Schema
	sub    router.BlockingSubscriptionPayload
}

func (s subscriptionHandler) do() *graphql.Result {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	ctx = context.WithValue(ctx, router.RawSubscriptionKey, true)
	params := graphql.Params{
		Schema:         *s.schema,
		RequestString:  s.sub.Context.Query,
		VariableValues: s.sub.Context.VariableValues,
		OperationName:  s.sub.Context.OperationName,
		Context:        ctx,
	}
	return graphql.Do(params)
}

func (s subscriptionHandler) writeResult(ws *websocket.Conn, result *graphql.Result) error {
	w, err := ws.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()
	enc := json.NewEncoder(w)
	if s.pretty {
		enc.SetIndent("", "\t")
	}
	if err := enc.Encode(result); err != nil {
		switch err.(type) {
		case *json.MarshalerError,
			*json.UnsupportedTypeError,
			*json.UnsupportedValueError:
			w.Write([]byte("ERROR: json marshal error: " + err.Error()))
		}
	}
	return err
}

// Handle subscription websocket
func (s subscriptionHandler) Handle(ws *websocket.Conn) {
	defer ws.Close()
	defer s.sub.Reader.Close()
	for s.sub.Reader.Next() {
		// execute graphql query
		if err := s.writeResult(ws, s.do()); err != nil {
			return
		}
	}
	if err := s.sub.Reader.Error(); err != nil {
		w, err := ws.NextWriter(websocket.TextMessage)
		if err != nil {
			klog.Error("unknown error", err)
			return
		}
		defer w.Close()
		w.Write([]byte("ERROR: " + err.Error()))
	}
}

// Handler implements http.Handler for GraphQL
type Handler struct {
	Schema   *graphql.Schema
	graphiql bool
	pretty   bool
	upgrader websocket.Upgrader
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	// get query
	opts := handler.NewRequestOptions(req)

	// execute graphql query
	params := graphql.Params{
		Schema:         *h.Schema,
		RequestString:  opts.Query,
		VariableValues: opts.Variables,
		OperationName:  opts.OperationName,
		Context:        ctx,
	}

	if h.graphiql && req.Method == http.MethodGet {
		acceptHeader := req.Header.Get("Accept")
		_, raw := req.URL.Query()["raw"]
		if !raw && !strings.Contains(acceptHeader, "application/json") && strings.Contains(acceptHeader, "text/html") {
			if websocket.IsWebSocketUpgrade(req) {
				http.Error(rw, "websocket not supported with GraphiQL", 400)
				return
			}
			renderGraphiQL(rw, params)
			return
		}
	}

	result := graphql.Do(params)
	if sub, ok := result.Extensions["subscriptionBlocking"].(router.BlockingSubscriptionPayload); ok && len(result.Errors) == 0 {
		conn, err := h.upgrader.Upgrade(rw, req, nil)
		if err != nil {
			klog.Error(err.Error())
			return
		}
		subHandler := subscriptionHandler{
			pretty: h.pretty,
			schema: h.Schema,
			sub:    sub,
		}
		subHandler.Handle(conn)
		return
	}

	rw.Header().Add("Content-Type", "application/json; charset=utf-8")
	var buff []byte
	rw.WriteHeader(http.StatusOK)
	if h.pretty {
		buff, _ = json.MarshalIndent(result, "", "\t")
	} else {
		buff, _ = json.Marshal(result)
	}
	rw.Write(buff)
}

// New returns new handler
func New(cfg Config) *Handler {
	h := Handler{
		Schema:   cfg.Schema,
		graphiql: cfg.GraphiQL,
		pretty:   cfg.Pretty,
		upgrader: websocket.Upgrader{
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			EnableCompression: true,
		},
	}
	return &h
}
