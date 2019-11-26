package types

type ResponsePath struct {
	Prev *ResponsePath `json:"responsePath,omitempty"`
	Key  string        `json:"key"`
}
