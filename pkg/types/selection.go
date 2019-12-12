package types

// FragmentDefinition is a fragment definition from client schema
type FragmentDefinition struct {
	Directives          Directives           `json:"directives,omitempty"`
	TypeCondition       TypeRef              `json:"typeCondition"`
	SelectionSet        Selections           `json:"selectionSet"`
	VariableDefinitions []VariableDefinition `json:"variableDefinitions,omitempty"`
}

// Selection is a represents a field or fragment requested by client
type Selection struct {
	Name         string              `json:"name,omitempty"`
	Arguments    Arguments           `json:"arguments,omitempty"`
	Directives   Directives          `json:"directives,omitempty"`
	SelectionSet Selections          `json:"selectionSet,omitempty"`
	Definition   *FragmentDefinition `json:"definition,omitempty"`
}

// Selections list of selections
type Selections []Selection
