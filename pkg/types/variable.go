package types

// Variable is a name of variable defined by client
type Variable struct {
	Name string `json:"name"`
}

// VariableDefinition client defined variable
type VariableDefinition struct {
	Variable     Variable    `json:"variable"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
}
