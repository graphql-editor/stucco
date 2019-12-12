package router_test

import (
	"testing"

	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentMerge(t *testing.T) {
	data := []struct {
		title    string
		in       router.Environment
		merge    router.Environment
		expected router.Environment
	}{
		{
			title: "OriginalStays",
			in: router.Environment{
				Provider: "provider1",
				Runtime:  "runtime1",
			},
			merge: router.Environment{
				Provider: "provider2",
				Runtime:  "runtime2",
			},
			expected: router.Environment{
				Provider: "provider1",
				Runtime:  "runtime1",
			},
		},
		{
			title: "Overrides",
			in: router.Environment{
				Provider: "",
				Runtime:  "",
			},
			merge: router.Environment{
				Provider: "provider2",
				Runtime:  "runtime2",
			},
			expected: router.Environment{
				Provider: "provider2",
				Runtime:  "runtime2",
			},
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			tt.in.Merge(tt.merge)
			assert.Equal(t, tt.expected, tt.in)
		})
	}
}
