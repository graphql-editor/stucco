package vars

import (
	global_vars "github.com/graphql-editor/stucco/pkg/vars"
)

// Vars meta variables relating to stucco itself
type Vars struct {
	global_vars.Vars
	AzureFunction string
}

// DefaultVars c
var DefaultVars = Vars{
	Vars: global_vars.DefaultVars,
}
