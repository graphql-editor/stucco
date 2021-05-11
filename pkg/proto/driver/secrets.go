package protodriver

import (
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
