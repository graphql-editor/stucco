package runtimes

import (
	"strings"

	"github.com/graphql-editor/stucco/pkg/providers/azure/project"
	"github.com/graphql-editor/stucco/pkg/types"
)

// StuccoJS runtime
type StuccoJS struct{}

// Function return stucco-js runtime function config
func (s StuccoJS) Function(f types.Function) (pf project.Function) {
	parts := strings.Split(f.Name, ".")
	pf.ScriptFile = parts[0] + ".js"
	if len(parts) == 2 && parts[1] != "js" {
		pf.EntryPoint = parts[1]
	}
	return
}

// LocalSettings are default local.settings.json for stucco-js
func (s StuccoJS) LocalSettings() (pl project.LocalSettings) {
	pl.WorkerRuntime = "stucco-js"
	return
}
