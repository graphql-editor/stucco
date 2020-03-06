package configs_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/graphql-editor/stucco/pkg/providers/azure/configs"
	"github.com/stretchr/testify/assert"
)

func TestLocalSettings(t *testing.T) {
	localSettingsJSON, err := ioutil.ReadFile("./testdata/local.settings.json")
	panicErr(err)
	var l configs.LocalSettings
	assert.NoError(t, json.Unmarshal(localSettingsJSON, &l))
	b, err := json.Marshal(l)
	assert.NoError(t, err)
	assert.JSONEq(t, string(localSettingsJSON), string(b))
}
