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
	re           = regexp.MustCompile(`^v?[0-9]*\.[0-9]*\.[0-9]*$`)
	Version      = func() string {
		v := BuildVersion
		if v == "" {
			v = BuildDate
		}
		if !re.Match([]byte(v)) {
			v = "dev-" + v
		}
		return v
	}()
)
