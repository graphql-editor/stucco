package types

// HttpRequest represents http request data.
type HttpRequest struct {
	Headers map[string]string `json:"headers,omitempty"`
}
