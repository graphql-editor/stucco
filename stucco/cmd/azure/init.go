/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package azurecmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/graphql-editor/stucco/pkg/providers/azure/configs"
	"github.com/graphql-editor/stucco/pkg/providers/azure/project"
	"github.com/graphql-editor/stucco/pkg/providers/azure/project/runtimes"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func checkNonExistentFile(p string) (err error) {
	_, err = os.Stat(p)
	if os.IsNotExist(err) {
		err = nil
	} else if err == nil {
		err = errors.Errorf("file %s exists", p)
	}
	return
}

func exitErr(err error, msg string) {
	klog.Fatalln(errors.Wrap(err, msg).Error())
}

const (
	initError = "could not initialize project"
)

type genOpts struct {
	localSettings, dockerfile, overwrite bool
	path, output, runtime, authLevel     string
}

func (g *genOpts) setupCommand(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&g.localSettings, "local-settings", "l", true, "Create local.settings.json")
	cmd.Flags().BoolVarP(&g.dockerfile, "dockerfile", "d", false, "Create boilerplate dockerfile for project")
	cmd.Flags().BoolVar(&g.overwrite, "overwrite", true, "Overwrite existing files")
	cmd.Flags().StringVarP(&g.path, "path", "p", g.path, "Project path")
	cmd.Flags().StringVarP(&g.output, "output", "o", filepath.Join(g.path, ".wwwroot"), "Output root path")
	cmd.Flags().StringVarP(&g.runtime, "runtime", "r", "stucco-js", "Stucco runtime name")
	cmd.Flags().StringVar(&g.authLevel, "auth-level", "function", "Function auth level")
}

func (g genOpts) projectRuntime() (r project.Runtime) {
	switch g.runtime {
	case "stucco-js":
		r = runtimes.StuccoJS{}
	}
	return
}

func isSubpath(p1, p2 string) bool {
	if !filepath.IsAbs(p2) {
		return true
	}
	p, err := filepath.Rel(p1, p2)
	return err == nil && strings.HasPrefix(
		filepath.Join(p1, p),
		filepath.Clean(p1),
	)
}

func (g genOpts) validate() error {
	switch g.runtime {
	case "stucco-js":
	default:
		return errors.Errorf("runtime %s is not a valid value", g.runtime)
	}
	switch configs.AuthLevel(g.authLevel) {
	case configs.AnonymousAuthLevel, configs.FunctionAuthLevel, configs.AdminAuthLevel:
	default:
		return errors.Errorf("auth level %s is not a valid value", g.authLevel)
	}
	if !isSubpath(g.path, g.output) {
		return errors.Errorf("output path %s is outside of %s", g.output, g.path)
	}
	return nil
}

func newGenOpts() genOpts {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return genOpts{
		localSettings: true,
		dockerfile:    false,
		overwrite:     true,
		path:          wd,
		output:        filepath.Join(wd, ".wwwroot"),
		runtime:       "stucco-js",
		authLevel:     "function",
	}
}

// NewInitCommand returns new init command
func NewInitCommand() *cobra.Command {
	genOpts := newGenOpts()
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initialize Azure Functions from stucco.json",
		Long:  `Creates Azure Functions confiugrations file based on your stucco.json.`,
		Run: func(cmd *cobra.Command, args []string) {
			hostJSON := filepath.Join(genOpts.output, "host.json")
			if err := checkNonExistentFile(hostJSON); err != nil {
				exitErr(err, initError)
			}
			localSettingsJSON := filepath.Join(genOpts.output, "local.settings.json")
			if err := checkNonExistentFile(localSettingsJSON); err != nil {
				exitErr(err, initError)
			}
			var cfg router.Config
			cfgPath := ""
			if genOpts.path != "" {
				cfgPath = filepath.Join(genOpts.path, "stucco")
			}
			if err := utils.LoadConfigFile(cfgPath, &cfg); err != nil {
				exitErr(err, initError)
			}
			if err := genOpts.validate(); err != nil {
				exitErr(err, initError)
			}
			p := project.Project{
				Config:             cfg,
				Output:             genOpts.output,
				Overwrite:          genOpts.overwrite,
				Path:               genOpts.path,
				Runtime:            genOpts.projectRuntime(),
				WriteLocalSettings: genOpts.localSettings,
				WriteDockerfile:    genOpts.dockerfile,
				AuthLevel:          configs.AuthLevel(genOpts.authLevel),
			}
			if err := p.Write(); err != nil {
				exitErr(err, initError)
			}
		},
	}
	genOpts.setupCommand(initCommand)
	return initCommand
}
