package configs_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/graphql-editor/stucco/pkg/providers/azure/configs"
	"github.com/stretchr/testify/assert"
)

func TestHost(t *testing.T) {
	hostJSON, err := ioutil.ReadFile("./testdata/host.json")
	panicErr(err)
	var h configs.Host
	assert.NoError(t, json.Unmarshal(hostJSON, &h))
	b, err := json.Marshal(h)
	assert.NoError(t, err)
	assert.JSONEq(t, string(hostJSON), string(b))
}
