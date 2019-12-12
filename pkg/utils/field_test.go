package utils_test

import (
	"testing"

	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestFieldName(t *testing.T) {
	assert.Equal(t, utils.FieldName("SomeType", "someField"), "SomeType.someField")
}
