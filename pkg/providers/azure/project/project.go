package project

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"log"

	"github.com/graphql-editor/stucco/pkg/providers/azure/configs"
	"github.com/graphql-editor/stucco/pkg/providers/azure/driver"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/pkg/errors"
)

const (
	functionJSONFileName      = "function.json"
	hostJSONFileName          = "host.json"
	localSettingsJSONFileName = "local.settings.json"
	dockerfileFilename        = "Dockerfile"
	filePerm                  = 0644
	dirPerm                   = 0755
	dockerfile                = `FROM gqleditor/stucco-js-azure-worker:node12

ENV AzureWebJobsScriptRoot=/home/site/wwwroot \
	AzureFunctionsJobHost__Logging__Console__IsEnabled=true

COPY {{ .Path }} /home/site/wwwroot
{{ if ne .Path .Output -}}
COPY %s/* /home/site/wwwroot/
{{- end }}
RUN cd /home/site/wwwroot && \
	npm install --production

WORKDIR /home/site/wwwroot`
)

var (
	defaultRoutePrefix = ""
	dockerfileTemplate = template.Must(template.New("Dockerfile").Parse(dockerfile))
)

// Function represents function data information used in generation of function.json
type Function struct {
	ScriptFile string
	EntryPoint string
}

// LocalSettings data used in generation of local.settings.json
type LocalSettings struct {
	WorkerRuntime string
	Values        map[string]string
}

// Runtime analyzes Stucco configuration
type Runtime interface {
	Function(f types.Function) Function
	LocalSettings() LocalSettings
}

// Project creates Azure Function configs from Stucco config
type Project struct {
	// Config is a Stucco configuration of a project
	Config router.Config
	// LocalSettingsValues is a map of values that should be written to local.settings.json
	LocalSettingsValues map[string]string
	// Output is a root path to which configs should be written
	Output string
	// Overwrite if set to true all existing files will be overwritten, otherwise they will be skipped.
	Overwrite bool
	// Path is projects path, by default it is current work directory
	Path string
	// Runtime of a project
	Runtime Runtime
	// WriteLocalSettings if set to true default local.settings.json will be written
	WriteLocalSettings bool
	// WriteDockerfile instructs project to generate boilerplate Dockerfile in project
	WriteDockerfile bool
}

func handlePath(path string, overwrite bool) (write bool, err error) {
	_, err = os.Stat(path)
	switch {
	case err == nil:
		if overwrite {
			err = os.RemoveAll(path)
			write = err == nil
		} else {
			log.Printf(fmt.Sprintf("skipping path %s because it exists", path))
		}
	case os.IsNotExist(err):
		err = nil
		write = true
	}
	return
}

func (p Project) writeFunctionConfig(path string, f types.Function) (err error) {
	var write bool
	if write, err = handlePath(path, p.Overwrite); err == nil && write {
		if err = os.Mkdir(path, dirPerm); err == nil {
			functionData := p.Runtime.Function(f)
			functionData.ScriptFile, err = filepath.Rel(
				path,
				filepath.Join(p.Path, functionData.ScriptFile),
			)
			if err == nil {
				var b []byte
				if b, err = json.Marshal(configs.Function{
					ScriptFile: functionData.ScriptFile,
					EntryPoint: functionData.EntryPoint,
					Bindings: []configs.Binding{
						configs.Binding{
							Name:      "req",
							Type:      configs.HTTPTrigger,
							Direction: configs.InDirection,
							Route:     driver.EndpointName(f.Name),
							AuthLevel: configs.FunctionAuthLevel,
							Methods:   []configs.Method{configs.PostMethod},
						},
						configs.Binding{
							Name:      "res",
							Type:      configs.HTTP,
							Direction: configs.OutDirection,
						},
					},
				}); err == nil {
					err = ioutil.WriteFile(filepath.Join(path, functionJSONFileName), b, filePerm)
				}
			}
		}
	}
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("function %s was not written", f.Name))
	}
	return
}

func filterUnique(m map[string]types.Function) map[string]types.Function {
	ff := make(map[types.Function]string, len(m))
	for k, v := range m {
		if f, ok := ff[v]; ok {
			ff[v] = f + "-" + k
		} else {
			ff[v] = k
		}
	}
	mm := make(map[string]types.Function, len(ff))
	for k, v := range ff {
		mm[v] = k
	}
	return mm
}

func (p Project) writeFunctions(path string) (err error) {
	functions := filterUnique(functionsFromConfig(p.Config))
	for fpath, f := range functions {
		if err = p.writeFunctionConfig(filepath.Join(path, fpath), f); err != nil {
			return
		}
	}
	return
}

func (p Project) writeHost(path string) error {
	path = filepath.Join(path, hostJSONFileName)
	b, err := json.Marshal(configs.Host{
		Version: "2.0",
		ExtensionBundle: &configs.ExtensionBundle{
			ID:      "Microsoft.Azure.Functions.ExtensionBundle",
			Version: "[1.*, 2.0.0)",
		},
		Extensions: &configs.Extensions{
			HTTP: &configs.HTTPExtension{
				RoutePrefix: &defaultRoutePrefix,
			},
		},
	})
	if err == nil {
		err = ioutil.WriteFile(path, b, filePerm)
	}
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("could not write %s", path))
	}
	return err
}

func (p Project) writeLocalSettings(path string) error {
	if !p.WriteLocalSettings {
		return nil
	}
	path = filepath.Join(path, localSettingsJSONFileName)
	values := make(map[string]string, len(p.LocalSettingsValues))
	for k, v := range p.LocalSettingsValues {
		values[k] = v
	}
	localSettings := p.Runtime.LocalSettings()
	if _, ok := values["FUNCTIONS_WORKER_RUNTIME"]; !ok {
		values["FUNCTIONS_WORKER_RUNTIME"] = localSettings.WorkerRuntime
	}
	for k, v := range localSettings.Values {
		if _, ok := values[k]; !ok {
			values[k] = v
		}
	}
	b, err := json.Marshal(configs.LocalSettings{
		Values: values,
	})
	if err == nil {
		err = ioutil.WriteFile(path, b, filePerm)
	}
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("could not write %s", path))
	}

	return nil
}

func (p Project) writeDockerfile(path string) (err error) {
	if !p.WriteDockerfile {
		return nil
	}
	path = filepath.Join(path, dockerfileFilename)
	defer func() {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("could not write %s", path))
		}
	}()
	f, err := os.Create(dockerfile)
	if err == nil {
		defer func() {
			ferr := f.Close()
			if err == nil {
				err = ferr
			}
		}()
		err = dockerfileTemplate.Execute(f, p)
	}
	return
}

func resolverName(p string) string {
	parts := strings.Split(p, ".")
	return "resolver-" + parts[0] + "-field-" + parts[1]
}

func interfaceName(p string) string {
	return "interface-" + p
}

func scalarParseName(p string) string {
	return "scalar-" + p + "-parse"
}

func scalarSerializeName(p string) string {
	return "scalar-" + p + "-serialize"
}

func unionName(p string) string {
	return "union-" + p
}

func functionsFromConfig(cfg router.Config) map[string]types.Function {
	functions := make(map[string]types.Function)
	for k, v := range cfg.Interfaces {
		functions[interfaceName(k)] = v.ResolveType
	}
	for k, v := range cfg.Resolvers {
		functions[resolverName(k)] = v.Resolve
	}
	for k, v := range cfg.Scalars {
		functions[scalarParseName(k)] = v.Parse
		functions[scalarSerializeName(k)] = v.Serialize
	}
	for k, v := range cfg.Unions {
		functions[unionName(k)] = v.ResolveType
	}
	return functions
}

func (p Project) out() (out string, err error) {
	out = p.Output
	if out == "" {
		out, err = os.Getwd()
	}
	_, err = os.Stat(out)
	if os.IsNotExist(err) {
		err = os.Mkdir(out, dirPerm)
	}
	return
}

// Write project from scratch
func (p Project) Write() (err error) {
	out, err := p.out()
	fmt.Println(out)
	if err == nil {
		err = p.writeFunctions(out)
	}
	if err == nil {
		err = p.writeHost(out)
	}
	if err == nil {
		err = p.writeLocalSettings(out)
	}
	if err == nil {
		err = p.writeDockerfile(out)
	}
	if err != nil {
		err = errors.Wrap(err, "could not write Azure Functions configs")
	}
	return
}

// Update project with functions
func (p Project) Update() (err error) {
	out, err := p.out()
	if err == nil {
		err = p.writeFunctions(out)
	}
	return err
}
