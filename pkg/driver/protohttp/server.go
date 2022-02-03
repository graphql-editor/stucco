package protohttp

import (
	"fmt"
	"io"
	"net/http"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
)

// Muxer for Protocol Buffer handler
type Muxer interface {
	Authorize(driver.AuthorizeInput) (bool, error)
	FieldResolve(driver.FieldResolveInput) (interface{}, error)
	InterfaceResolveType(driver.InterfaceResolveTypeInput) (string, error)
	SetSecrets(driver.SetSecretsInput) error
	ScalarParse(driver.ScalarParseInput) (interface{}, error)
	ScalarSerialize(driver.ScalarSerializeInput) (interface{}, error)
	UnionResolveType(driver.UnionResolveTypeInput) (string, error)
	SubscriptionConnection(driver.SubscriptionConnectionInput) (interface{}, error)
}

// ErrorLogger logs unrecoverable errors while handling request
type ErrorLogger interface {
	Error(err error)
}

// Handler is a http.Handler for Protocol Buffers server
type Handler struct {
	Muxer
	ErrorLogger
}

type httpError interface {
	Write(http.ResponseWriter)
}

type requestError struct {
	msg    string
	args   []interface{}
	status int
}

func (e requestError) Write(rw http.ResponseWriter) {
	rw.Header().Add(contentTypeHeader, "text/plain")
	fmt.Fprintf(rw, e.msg, e.args...)
	rw.WriteHeader(e.status)
}

type badRequest struct {
	msg  string
	args []interface{}
}

func (e badRequest) Write(rw http.ResponseWriter) {
	requestError{
		msg:    "BadRequest: " + e.msg,
		args:   e.args,
		status: http.StatusBadRequest,
	}.Write(rw)
}

type internalServerError struct {
	msg  string
	args []interface{}
}

func (e internalServerError) Write(rw http.ResponseWriter) {
	requestError{
		msg:    "InternalServerError: " + e.msg,
		args:   e.args,
		status: http.StatusInternalServerError,
	}.Write(rw)
}

func (h *Handler) serveHTTP(req *http.Request, rw http.ResponseWriter) error {
	messageType, err := getMessageType(req.Header.Get(contentTypeHeader))
	if err != nil {
		br := badRequest{
			msg:  "invalid content type: %s",
			args: []interface{}{err.Error()},
		}
		br.Write(rw)
		return nil
	}
	switch messageType {
	case string(authorizeRequestMessage):
		err = h.authorize(req, rw)
	case string(fieldResolveRequestMessage):
		err = h.fieldResolve(req, rw)
	case string(interfaceResolveTypeRequestMessage):
		err = h.interfaceResolveType(req, rw)
	case string(setSecretsRequestMessage):
		err = h.setSecrets(req, rw)
	case string(scalarParseRequestMessage):
		err = h.scalarParse(req, rw)
	case string(scalarSerializeRequestMessage):
		err = h.scalarSerialize(req, rw)
	case string(unionResolveTypeRequestMessage):
		err = h.unionResolveType(req, rw)
	case string(subscriptionConnectionRequestMessage):
		err = h.subscriptionConnection(req, rw)
	default:
		br := badRequest{
			msg:  "invalid content type: %s",
			args: []interface{}{err.Error()},
		}
		br.Write(rw)
		return nil
	}
	return err
}

func writeProto(w io.Writer, p protobuf.Message) error {
	b, err := protobuf.Marshal(p)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}

// ServeHTTP implements http.Handler interface for Protocol Buffer server
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := h.serveHTTP(req, rw); h.ErrorLogger != nil {
		h.ErrorLogger.Error(err)
	}
}
