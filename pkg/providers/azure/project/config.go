package project

import (
	"encoding/json"

	"github.com/graphql-editor/stucco/pkg/server"
)

type AzureOpts struct {
	Webhooks []string `json:"webhooks"`
}

// Config represents azure server config
type Config struct {
	server.Config
	AzureOpts AzureOpts `json:"azureOpts"`
}

func (c *Config) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &c.Config); err != nil {
		return err
	}
	opts := struct {
		AzureOpts AzureOpts `json:"azureOpts"`
	}{}
	err := json.Unmarshal(data, &opts)
	if err == nil {
		c.AzureOpts = opts.AzureOpts
	}
	return err
}
