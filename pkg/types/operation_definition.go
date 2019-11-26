package types

type OperationDefinition struct {
	Operation           string               `json:"operation"`
	Name                string               `json:"name"`
	VariableDefinitions []VariableDefinition `json:"variableDefinitions,omitempty"`
	Directives          Directives           `json:"directives,omitempty"`
	SelectionSet        Selections           `json:"selectionSet,omitempty"`
}
