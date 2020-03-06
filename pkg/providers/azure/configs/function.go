// Package azureconfigs contains structs to representing Azure Function configs
package configs

// ConfigurationSource enum.
type ConfigurationSource string

const (
	// AttributesConfigurationSource use WebJobs attributes
	AttributesConfigurationSource ConfigurationSource = "attributes"
	// ConfigConfigurationSource use bindings from this configuration
	ConfigConfigurationSource ConfigurationSource = "config"
)

// Function represents azure function configuration
type Function struct {
	Disabled            *bool               `json:"disabled,omitempty"`
	Excluded            *bool               `json:"excluded,omitempty"`
	ScriptFile          string              `json:"scriptFile,omitempty"`
	EntryPoint          string              `json:"entryPoint,omitempty"`
	ConfigurationSource ConfigurationSource `json:"configurationSource,omitempty"`
	Bindings            []Binding           `json:"bindings,omitempty"`
}
