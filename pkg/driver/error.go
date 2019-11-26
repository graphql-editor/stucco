package driver

// Error passed between runner and router
type Error struct {
	Message string `json:"message,omitempty"`
}
