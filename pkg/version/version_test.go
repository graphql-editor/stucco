package version_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersions(t *testing.T) {
	o, err := exec.Command("go", "run", "testdata/main.go").Output()
	assert.NoError(t, err)
	assert.Regexp(t, "^dev-[0-9]{12}$", string(o))
	o, err = exec.Command("go", "run", "-ldflags=-X github.com/graphql-editor/stucco/pkg/version.BuildVersion=v1.0.0", "testdata/main.go").Output()
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.0", string(o))
	o, err = exec.Command("go", "run", "-ldflags=-X github.com/graphql-editor/stucco/pkg/version.BuildVersion=123456abcdef", "testdata/main.go").Output()
	assert.NoError(t, err)
	assert.Equal(t, "dev-123456abcdef", string(o))
}
