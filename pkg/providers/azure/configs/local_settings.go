package configs

// LocalSettingsHost is host property in local.settings.json
type LocalSettingsHost struct {
	LocalHTTPPort   int    `json:"LocalHTTPPort,omitempty"`
	CORS            string `json:"CORS,omitempty"`
	CORSCredentials *bool  `json:"CORSCredentials,omitempty"`
}

// LocalSettings is local.settings.json
type LocalSettings struct {
	IsEncrypted       bool               `json:"IsEncrypted"`
	Values            map[string]string  `json:"Values"`
	Host              *LocalSettingsHost `json:"Host,omitempty"`
	ConnectionStrings map[string]string  `json:"ConnectionStrings,omitempty"`
}
