package types

type TypeRef struct {
	Name    string   `json:"name,omitempty"`
	NonNull *TypeRef `json:"nonNull,omitempty"`
	List    *TypeRef `json:"list,omitempty"`
}
