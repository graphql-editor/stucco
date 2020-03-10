package utils_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/stretchr/testify/assert"
)

var expectedConfig = func() router.Config {
	var cfg router.Config
	b, err := ioutil.ReadFile("./testdata/config.json")
	if err == nil {
		err = json.Unmarshal(b, &cfg)
	}
	if err != nil {
		panic(err)
	}
	return cfg
}

func TestLoadFileFromFileSystem(t *testing.T) {
	var cfg router.Config
	assert.NoError(t, utils.LoadConfigFile("./testdata/config", &cfg))
	assert.Equal(t, expectedConfig(), cfg)
	assert.Error(t, utils.LoadConfigFile("./testdata/invalid", &cfg))
}

func TestLoadFileFromRemote(t *testing.T) {
	srv := http.Server{
		Handler: http.FileServer(http.Dir("./testdata")),
	}
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	go srv.Serve(l)
	defer srv.Shutdown(context.Background())
	var cfg router.Config
	assert.NoError(t, utils.LoadConfigFile("http://localhost:8080/config.json", &cfg))
	assert.Equal(t, expectedConfig(), cfg)
	assert.Error(t, utils.LoadConfigFile("http://localhost/config", &cfg))
	assert.NoError(t, utils.LoadConfigFile("http://localhost:8080/config.json?some=arg&in=url", &cfg))
	assert.Equal(t, expectedConfig(), cfg)
}
