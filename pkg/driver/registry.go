package driver

import (
	"sync"
)

type Config struct {
	Provider string `json:"provider,omitempty"`
	Runtime  string `json:"runtime,omitempty"`
}

var (
	lock    = sync.Mutex{}
	drivers = map[Config]Driver{}
)

// Register adds a new driver
func Register(c Config, d Driver) {
	lock.Lock()
	drivers[c] = d
	lock.Unlock()
}

func GetDriver(c Config) Driver {
	lock.Lock()
	d := drivers[c]
	lock.Unlock()
	return d
}
