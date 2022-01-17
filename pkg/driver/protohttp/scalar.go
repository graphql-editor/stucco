package protohttp

import (
	"bytes"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// ScalarParse over http
func (c *Client) ScalarParse(input driver.ScalarParseInput) driver.ScalarParseOutput {
	var out driver.ScalarParseOutput
	var body bytes.Buffer
	err := protodriver.WriteScalarParseInput(&body, input)
	if err == nil {
		var b []byte
		if b, err = c.do(message{
			contentType: fieldResolveRequestMessage,
			b:           body.Bytes(),
		}); err == nil {
			out, err = protodriver.ReadScalarParseOutput(bytes.NewReader(b))
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

// ScalarSerialize over http
func (c *Client) ScalarSerialize(input driver.ScalarSerializeInput) driver.ScalarSerializeOutput {
	var out driver.ScalarSerializeOutput
	var body bytes.Buffer
	err := protodriver.WriteScalarSerializeInput(&body, input)
	if err == nil {
		var b []byte
		if b, err = c.do(message{
			contentType: fieldResolveRequestMessage,
			b:           body.Bytes(),
		}); err == nil {
			out, err = protodriver.ReadScalarSerializeOutput(bytes.NewReader(b))
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) scalarParse(req *http.Request, rw http.ResponseWriter) error {
	rw.Header().Add(contentTypeHeader, scalarParseResponseMessage.String())
	in, err := protodriver.ReadScalarParseInput(req.Body)
	if err == nil {
		req.Body.Close()
		if err == nil {
			var driverResp interface{}
			driverResp, err = h.ScalarParse(in)
			if err == nil {
				err = protodriver.WriteScalarParseOutput(rw, driverResp)
			}
		}
	}
	if err != nil {
		err = writeProto(rw, &protoMessages.ScalarParseResponse{
			Error: &protoMessages.Error{
				Msg: err.Error(),
			},
		})
	}
	return err
}

func (h *Handler) scalarSerialize(req *http.Request, rw http.ResponseWriter) error {
	rw.Header().Add(contentTypeHeader, scalarSerializeResponseMessage.String())
	in, err := protodriver.ReadScalarSerializeInput(req.Body)
	if err == nil {
		req.Body.Close()
		if err == nil {
			var driverResp interface{}
			driverResp, err = h.ScalarSerialize(in)
			if err == nil {
				err = protodriver.WriteScalarSerializeOutput(rw, driverResp)
			}
		}
	}
	if err != nil {
		err = writeProto(rw, &protoMessages.ScalarSerializeResponse{
			Error: &protoMessages.Error{
				Msg: err.Error(),
			},
		})
	}
	return err
}
