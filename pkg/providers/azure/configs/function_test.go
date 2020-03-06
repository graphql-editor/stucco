package configs_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/graphql-editor/stucco/pkg/providers/azure/configs"
	"github.com/stretchr/testify/assert"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestFunction(t *testing.T) {
	functionJSON, err := ioutil.ReadFile("./testdata/function.json")
	panicErr(err)
	var f configs.Function
	assert.NoError(t, json.Unmarshal(functionJSON, &f))
	b, err := json.Marshal(f)
	assert.NoError(t, err)
	assert.JSONEq(t, string(functionJSON), string(b))
	assert.True(t, f.Bindings[0].Type.Dynamic())
	for _, binding := range f.Bindings[1:] {
		assert.False(t, binding.Type.Dynamic())
	}
}
