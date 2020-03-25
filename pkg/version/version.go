package version

import (
	"regexp"
	"time"
)

// BuildVersion is set on compile time representing build version
// If empty, build version was not set
var (
	BuildVersion string
	BuildDate    = time.Now().UTC().Format("200601021504")
)

// Version represents stucco version it can be either a release version or dev
var Version = func() string {
	if BuildVersion == "" {
		return "dev-" + BuildDate
	}
	release, err := regexp.Match(`^v?[0-9]*\.[0-9]*\.[0-9]*$`, []byte(BuildVersion))
	if err != nil {
		panic(err)
	}
	if release {
		return BuildVersion
	}
	return "dev-" + BuildVersion
}()
