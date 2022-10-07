package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
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
	RouterConfig router.Config
	Schema       *graphql.Schema
	Pretty       bool
	GraphiQL     bool
	RootObjectFn handler.RootObjectFn
	CheckOrigin  func(req *http.Request) bool
}

// subscriptionHandler is a websocket handler
type subscriptionHandler struct {
	pretty     bool
	schema     *graphql.Schema
	sub        router.BlockingSubscriptionPayload
	ctx        context.Context
	rootObject map[string]interface{}
}

func (s subscriptionHandler) do(v interface{}) *graphql.Result {
	ctx, cancel := context.WithTimeout(s.ctx, time.Second*30)
	defer cancel()
	ctx = context.WithValue(ctx, router.RawSubscriptionKey, true)
	ctx = context.WithValue(ctx, router.SubscriptionPayloadKey, v)
	params := graphql.Params{
		Schema:         *s.schema,
		RequestString:  s.sub.Context.Query,
		VariableValues: s.sub.Context.VariableValues,
		OperationName:  s.sub.Context.OperationName,
		Context:        ctx,
		RootObject:     s.rootObject,
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
		v, err := s.sub.Reader.Read()
		if err != nil {
			klog.Error("unknown error", err)
			return
		}
		// execute graphql query
		if err := s.writeResult(ws, s.do(v)); err != nil {
			klog.Error("unknown error", err)
			return
		}
	}
	if err := s.sub.Reader.Error(); err != nil {
		w, nerr := ws.NextWriter(websocket.TextMessage)
		if nerr != nil {
			klog.Error("unknown error", nerr)
			return
		}
		defer w.Close()
		w.Write([]byte("ERROR: " + err.Error()))
	}
}

// Handler implements http.Handler for GraphQL
type Handler struct {
	Schema       *graphql.Schema
	graphiql     bool
	pretty       bool
	upgrader     websocket.Upgrader
	rootObjectFn handler.RootObjectFn
}

type requestOptions struct {
	Query               string                 `json:"query" url:"query" schema:"query"`
	Variables           map[string]interface{} `json:"variables" url:"variables" schema:"variables"`
	OperationName       string                 `json:"operationName" url:"operationName" schema:"operationName"`
	RawSubscription     bool                   `json:"rawSubscription" url:"rawSubscription" schema:"rawSubscription"`
	SubscriptionPayload string                 `json:"subscriptionPayload" url:"subscriptionPayload" schema:"subscriptionPayload"`
}

// a workaround for getting`variables` as a JSON string
type requestOptionsCompatibility struct {
	Query               string `json:"query" url:"query" schema:"query"`
	Variables           string `json:"variables" url:"variables" schema:"variables"`
	OperationName       string `json:"operationName" url:"operationName" schema:"operationName"`
	RawSubscription     bool   `json:"rawSubscription" url:"rawSubscription" schema:"rawSubscription"`
	SubscriptionPayload string `json:"subscriptionPayload" url:"subscriptionPayload" schema:"subscriptionPayload"`
}

func valueBool(v url.Values, k string) bool {
	p, ok := v[k]
	if !ok || len(p) == 0 {
		return ok && len(p) == 0
	}
	switch p[0] {
	case "", "1", "true":
		return true
	}
	return false
}

func getFromForm(values url.Values) *requestOptions {
	query := values.Get("query")
	if query != "" {
		// get variables map
		variables := make(map[string]interface{}, len(values))
		variablesStr := values.Get("variables")
		json.Unmarshal([]byte(variablesStr), &variables)
		return &requestOptions{
			Query:               query,
			Variables:           variables,
			OperationName:       values.Get("operationName"),
			RawSubscription:     valueBool(values, "raw"),
			SubscriptionPayload: values.Get("subscriptionPayload"),
		}
	}

	return nil
}

func newRequestOptions(r *http.Request) *requestOptions {
	if reqOpt := getFromForm(r.URL.Query()); reqOpt != nil {
		return reqOpt
	}

	if r.Method != http.MethodPost {
		return &requestOptions{}
	}

	if r.Body == nil {
		return &requestOptions{}
	}

	// TODO: improve Content-Type handling
	contentTypeStr := r.Header.Get("Content-Type")
	contentTypeTokens := strings.Split(contentTypeStr, ";")
	contentType := contentTypeTokens[0]

	switch contentType {
	case handler.ContentTypeGraphQL:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return &requestOptions{}
		}
		return &requestOptions{
			Query: string(body),
		}
	case handler.ContentTypeFormURLEncoded:
		if err := r.ParseForm(); err != nil {
			return &requestOptions{}
		}

		if reqOpt := getFromForm(r.PostForm); reqOpt != nil {
			return reqOpt
		}

		return &requestOptions{}

	case handler.ContentTypeJSON:
		fallthrough
	default:
		var opts requestOptions
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return &opts
		}
		err = json.Unmarshal(body, &opts)
		if err != nil {
			// Probably `variables` was sent as a string instead of an object.
			// So, we try to be polite and try to parse that as a JSON string
			var optsCompatible requestOptionsCompatibility
			json.Unmarshal(body, &optsCompatible)
			json.Unmarshal([]byte(optsCompatible.Variables), &opts.Variables)
		}
		return &opts
	}
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	// get query
	opts := newRequestOptions(req)
	if opts.RawSubscription {
		ctx = context.WithValue(ctx, router.RawSubscriptionKey, true)
	}
	if opts.SubscriptionPayload != "" {
		ctx = context.WithValue(ctx, router.SubscriptionPayloadKey, opts.SubscriptionPayload)
	}

	// execute graphql query
	params := graphql.Params{
		Schema:         *h.Schema,
		RequestString:  opts.Query,
		VariableValues: opts.Variables,
		OperationName:  opts.OperationName,
		Context:        ctx,
	}
	if h.rootObjectFn != nil {
		params.RootObject = h.rootObjectFn(ctx, req)
	}

	if h.graphiql && req.Method == http.MethodGet {
		acceptHeader := req.Header.Get("Accept")
		if !opts.RawSubscription && !strings.Contains(acceptHeader, "application/json") && strings.Contains(acceptHeader, "text/html") {
			if websocket.IsWebSocketUpgrade(req) {
				http.Error(rw, "websocket not supported with GraphiQL", 400)
				return
			}
			renderGraphiQL(rw, params)
			return
		}
	}

	pctx, cancel := context.WithTimeout(ctx, time.Second*30)
	params.Context = pctx
	result := graphql.Do(params)
	cancel()
	if sub, ok := result.Extensions["subscriptionBlocking"].(router.BlockingSubscriptionPayload); ok && len(result.Errors) == 0 {
		defer func() {
			if r := recover(); r != nil {
				klog.Error(r)
			}
		}()
		conn, err := h.upgrader.Upgrade(rw, req, nil)
		if err != nil {
			klog.Error(err.Error())
			return
		}
		subHandler := subscriptionHandler{
			pretty:     h.pretty,
			schema:     h.Schema,
			sub:        sub,
			ctx:        ctx,
			rootObject: params.RootObject,
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

type webhookResponseWrapper struct {
	bytes.Buffer
	http.ResponseWriter
	status int
}

func (w *webhookResponseWrapper) Write(b []byte) (int, error) {
	return w.Buffer.Write(b)
}

func (w *webhookResponseWrapper) WriteHeader(status int) {
	w.status = status
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
		rootObjectFn: cfg.RootObjectFn,
	}
	if cfg.CheckOrigin != nil {
		h.upgrader.CheckOrigin = cfg.CheckOrigin
	}
	return &h
}
