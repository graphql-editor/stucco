package types

type FragmentDefinition struct {
	Directives          Directives           `json:"directives,omitempty"`
	TypeCondition       TypeRef              `json:"typeCondition"`
	SelectionSet        Selections           `json:"selectionSet"`
	VariableDefinitions []VariableDefinition `json:"variableDefinitions,omitempty"`
}

type Selection struct {
	Name         string              `json:"name,omitempty"`
	Arguments    Arguments           `json:"arguments,omitempty"`
	Directives   Directives          `json:"directives,omitempty"`
	SelectionSet Selections          `json:"selectionSet,omitempty"`
	Definition   *FragmentDefinition `json:"definition,omitempty"`
}

type Selections []Selection
