package vars

// Release contains deployment configs
type Release struct {
	Host       string
	DevVersion string
}

// Vars meta variables relating to stucco itself
type Vars struct {
	Relase Release
}

// DefaultVars c
var DefaultVars = Vars{
	Relase: Release{
		Host:       "stucco-release.fra1.cdn.digitaloceanspaces.com",
		DevVersion: "latest",
	},
}
