package types

type Directive struct {
	Name      string    `json:"name"`
	Arguments Arguments `json:"arguments,omitempty"`
}

type Directives []Directive
