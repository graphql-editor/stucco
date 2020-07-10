package protohttp

import (
	"fmt"
	"net/http"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
)

// Muxer for Protocol Buffer handler
type Muxer interface {
	FieldResolve(driver.FieldResolveInput) (interface{}, error)
	InterfaceResolveType(driver.InterfaceResolveTypeInput) (string, error)
	SetSecrets(driver.SetSecretsInput) error
	ScalarParse(driver.ScalarParseInput) (interface{}, error)
	ScalarSerialize(driver.ScalarSerializeInput) (interface{}, error)
	UnionResolveType(driver.UnionResolveTypeInput) (string, error)
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

func (h *Handler) serveHTTP(req *http.Request) (
	b []byte,
	responseContent protobufMessageContentType,
	herr httpError,
) {
	messageType, err := getMessageType(req.Header.Get(contentTypeHeader))
	if err != nil {
		herr = &badRequest{
			msg:  "invalid content type: %s",
			args: []interface{}{err.Error()},
		}
		return
	}
	var response protobuf.Message
	switch messageType {
	case string(fieldResolveRequestMessage):
		responseContent = fieldResolveResponseMessage
		response = h.fieldResolve(req)
	case string(interfaceResolveTypeRequestMessage):
		responseContent = interfaceResolveTypeResponseMessage
		response = h.interfaceResolveType(req)
	case string(setSecretsRequestMessage):
		responseContent = setSecretsResponseMessage
		response = h.setSecrets(req)
	case string(scalarParseRequestMessage):
		responseContent = scalarParseResponseMessage
		response = h.scalarParse(req)
	case string(scalarSerializeRequestMessage):
		responseContent = scalarSerializeResponseMessage
		response = h.scalarSerialize(req)
	case string(unionResolveTypeRequestMessage):
		responseContent = unionResolveTypeResponseMessage
		response = h.unionResolveType(req)
	default:
		herr = &badRequest{
			msg:  "invalid protobuf message type: %s",
			args: []interface{}{messageType},
		}
		return
	}
	b, err = protobuf.Marshal(response)
	if err != nil {
		herr = &internalServerError{
			msg:  "%s",
			args: []interface{}{err.Error()},
		}
	}
	return
}

// ServeHTTP implements http.Handler interface for Protocol Buffer server
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	response, responseContent, err := h.serveHTTP(req)
	if err != nil {
		err.Write(rw)
		return
	}
	rw.Header().Add(contentTypeHeader, responseContent.String())
	rw.WriteHeader(http.StatusOK)
	if _, err := rw.Write(response); err != nil && h.ErrorLogger != nil {
		h.ErrorLogger.Error(err)
	}
}
