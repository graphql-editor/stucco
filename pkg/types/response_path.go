package types

// ResponsePath is a node in response path.
type ResponsePath struct {
	Prev *ResponsePath `json:"responsePath,omitempty"`
	Key  interface{}   `json:"key"`
}
