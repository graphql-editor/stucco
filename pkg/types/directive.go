package types

// Directive represents an applied directive with arguments
type Directive struct {
	Name      string    `json:"name"`
	Arguments Arguments `json:"arguments,omitempty"`
}

// Directives is a list of directives
type Directives []Directive
