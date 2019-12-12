package driver

import (
	"sync"
)

// Config defines a config that a driver satisfies.
// Only on driver per config can be definied in registry
type Config struct {
	Provider string `json:"provider,omitempty"`
	Runtime  string `json:"runtime,omitempty"`
}

var (
	lock    = sync.Mutex{}
	drivers = map[Config]Driver{}
)

// Register adds a new driver for a user config
func Register(c Config, d Driver) {
	lock.Lock()
	drivers[c] = d
	lock.Unlock()
}

// GetDriver returns a driver matching user config for a runner
func GetDriver(c Config) Driver {
	lock.Lock()
	d := drivers[c]
	lock.Unlock()
	return d
}
