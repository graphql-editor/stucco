package protohttp_test

import (
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/protohttp"
	"github.com/stretchr/testify/assert"
)

func TestClientImplementsDriver(t *testing.T) {
	assert.Implements(t, (*driver.Driver)(nil), new(protohttp.Client))
}
