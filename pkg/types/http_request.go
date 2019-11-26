package types

type HttpRequest struct {
	Headers map[string]string `json:"headers,omitempty"`
}
