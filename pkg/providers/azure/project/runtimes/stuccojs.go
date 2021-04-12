package runtimes

import (
	"bytes"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/kennygrant/sanitize"
)

var stuccoJSFunctionJSONTemplate = template.Must(template.New("function.json").Parse(`{
	"bindings": [
		{
			"authLevel": "anonymous",
			"type": "httpTrigger",
			"direction": "in",
			"name": "req",
			"route": "{{ .route }}",
			"methods": [
				"get",
				"post"
			]
		},
		{
			"type": "http",
			"direction": "out",
			"name": "res"
		}
	]
}`))

// StuccoJS runtime
type StuccoJS struct {
	OsType       OsType
	MajorVersion int
}

var baseNameSeparators = regexp.MustCompile(`[./\\]`)

// Function return stucco-js runtime function config
func (s StuccoJS) Function(f types.Function) (files []File, err error) {
	parts := strings.Split(filepath.Clean(f.Name), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	var buf bytes.Buffer
	if err = stuccoJSFunctionJSONTemplate.Execute(&buf, map[string]interface{}{"route": strings.Join(parts, "/")}); err == nil {
		files = append(files, File{
			Reader: bytes.NewReader(buf.Bytes()),
			Path: filepath.Join(
				sanitize.BaseName(baseNameSeparators.ReplaceAllString(f.Name, "-")),
				"function.json",
			),
		})
	}
	return
}

// IgnoreFiles returns a list of Glob patterns to be ignored while creating runtime bundle
func (s StuccoJS) IgnoreFiles() []string {
	return append([]string{"/dist/*"}, commonIgnoreList...)
}

// GlobalFiles returns shared config files for runtime
func (s StuccoJS) GlobalFiles() ([]File, error) {
	return []File{
		{
			Reader: strings.NewReader(`{
				  "version": "2.0",
				  "logging": {
					"applicationInsights": {
					  "samplingSettings": {
						"isEnabled": true,
						"excludedTypes": "Request"
					  }
					},
					"logLevel": {
					  "default": "Information"
					}
				  },
				  "extensionBundle": {
					"id": "Microsoft.Azure.Functions.ExtensionBundle",
					"version": "[1.*, 2.0.0)"
				  },
				  "customHandler": {
					"description": {
					  "defaultExecutablePath": "node",
					  "arguments": ["./node_modules/stucco-js/lib/cli/cli.js", "azure", "serve"]
					},
					"enableForwardingHttpRequest": true
				  },
				  "extensions": {"http": {"routePrefix": ""}}
				}`),
			Path: "host.json",
		},
	}, nil
}
