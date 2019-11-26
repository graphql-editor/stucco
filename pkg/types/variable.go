package types

type Variable struct {
	Name string `json:"name"`
}

type VariableDefinition struct {
	Variable     Variable    `json:"variable"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
}
