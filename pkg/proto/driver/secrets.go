package protodriver

import (
	"io"
	"io/ioutil"

	protobuf "google.golang.org/protobuf/proto"

	"github.com/graphql-editor/stucco/pkg/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// MakeSetSecretsRequest creates protoMessages.SetSecretsRequest from driver.SetSecretsInput
func MakeSetSecretsRequest(input driver.SetSecretsInput) *protoMessages.SetSecretsRequest {
	s := new(protoMessages.SetSecretsRequest)
	for k, v := range input.Secrets {
		s.Secrets = append(s.Secrets, &protoMessages.Secret{
			Key:   k,
			Value: v,
		})
	}
	return s
}

// MakeSetSecretsResponse creates protoMessages.SetSecretsResponse from error
func MakeSetSecretsResponse(err error) *protoMessages.SetSecretsResponse {
	s := new(protoMessages.SetSecretsResponse)
	if err != nil {
		s.Error = &protoMessages.Error{
			Msg: err.Error(),
		}
	}
	return s
}

// MakeSetSecretsInput creates driver.SetSecretsInput from protoMessages.SetSecretsRequest
func MakeSetSecretsInput(req *protoMessages.SetSecretsRequest) driver.SetSecretsInput {
	var in driver.SetSecretsInput
	secrets := req.GetSecrets()
	if len(secrets) > 0 {
		in.Secrets = make(driver.Secrets, len(secrets))
		for _, v := range secrets {
			in.Secrets[v.Key] = v.Value
		}
	}
	return in
}

// MakeSetSecretsOutput creates driver.SetSecretsOutput from protoMessages.SetSecretsResponse
func MakeSetSecretsOutput(resp *protoMessages.SetSecretsResponse) driver.SetSecretsOutput {
	var out driver.SetSecretsOutput
	if resp.GetError() != nil {
		out.Error = &driver.Error{
			Message: resp.GetError().GetMsg(),
		}
	}
	return out
}

// ReadSetSecretsInput reads io.Reader until io.EOF and returs driver.SetSecretsInput
func ReadSetSecretsInput(r io.Reader) (driver.SetSecretsInput, error) {
	var err error
	var b []byte
	var out driver.SetSecretsInput
	protoMsg := new(protoMessages.SetSecretsRequest)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out = MakeSetSecretsInput(protoMsg)
		}
	}
	return out, err
}

// WriteSetSecretsInput writes SetSecretsInput into io.Writer
func WriteSetSecretsInput(w io.Writer, input driver.SetSecretsInput) error {
	req := MakeSetSecretsRequest(input)
	b, err := protobuf.Marshal(req)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}

// ReadSetSecretsOutput reads io.Reader until io.EOF and returs driver.SetSecretsOutput
func ReadSetSecretsOutput(r io.Reader) (driver.SetSecretsOutput, error) {
	var err error
	var b []byte
	var out driver.SetSecretsOutput
	protoMsg := new(protoMessages.SetSecretsResponse)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out = MakeSetSecretsOutput(protoMsg)
		}
	}
	return out, err
}

// WriteSetSecretsOutput writes SetSecretsOutput into io.Writer
func WriteSetSecretsOutput(w io.Writer, rerr error) error {
	req := MakeSetSecretsResponse(rerr)
	b, err := protobuf.Marshal(req)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}
