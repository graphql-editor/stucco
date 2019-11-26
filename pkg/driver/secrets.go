package driver

type Secrets map[string]string

type SetSecretsInput struct {
	// Secrets is a map of references which driver uses to populate secrets map
	Secrets Secrets
}

type SetSecretsOutput struct {
	Error *Error `json:"error,omitempty"`
}
