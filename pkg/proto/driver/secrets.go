package protodriver

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
)

// MakeSetSecretsRequest creates proto.SetSecretsRequest from driver.SetSecretsInput
func MakeSetSecretsRequest(input driver.SetSecretsInput) *proto.SetSecretsRequest {
	s := new(proto.SetSecretsRequest)
	for k, v := range input.Secrets {
		s.Secrets = append(s.Secrets, &proto.Secret{
			Key:   k,
			Value: v,
		})
	}
	return s
}

// MakeSetSecretsResponse creates proto.SetSecretsResponse from error
func MakeSetSecretsResponse(err error) *proto.SetSecretsResponse {
	s := new(proto.SetSecretsResponse)
	if err != nil {
		s.Error = &proto.Error{
			Msg: err.Error(),
		}
	}
	return s
}

// MakeSetSecretsInput creates driver.SetSecretsInput from proto.SetSecretsRequest
func MakeSetSecretsInput(req *proto.SetSecretsRequest) driver.SetSecretsInput {
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

// MakeSetSecretsOutput creates driver.SetSecretsOutput from proto.SetSecretsResponse
func MakeSetSecretsOutput(resp *proto.SetSecretsResponse) driver.SetSecretsOutput {
	var out driver.SetSecretsOutput
	if resp.GetError() != nil {
		out.Error = &driver.Error{
			Message: resp.GetError().GetMsg(),
		}
	}
	return out
}
