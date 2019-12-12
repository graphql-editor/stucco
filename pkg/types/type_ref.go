package types

// TypeRef is a reference to a type defined in schema
type TypeRef struct {
	Name    string   `json:"name,omitempty"`
	NonNull *TypeRef `json:"nonNull,omitempty"`
	List    *TypeRef `json:"list,omitempty"`
}
