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
package cmd

import (
	"os"
	"path/filepath"

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

// azureInitCmd represents the init command
var (
	wd = func() string {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		return wd
	}()
	localSettings = true
	overwrite     = true
	path          = wd
	output        = wd
	runtime       = ""

	azureInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize Azure Functions from stucco.json",
		Long:  `Creates Azure Functions confiugrations file based on your stucco.json.`,
		Run: func(cmd *cobra.Command, args []string) {
			hostJSON := filepath.Join(output, "host.json")
			if err := checkNonExistentFile(hostJSON); err != nil {
				exitErr(err, initError)
			}
			localSettingsJSON := filepath.Join(output, "local.settings.json")
			if err := checkNonExistentFile(localSettingsJSON); err != nil {
				exitErr(err, initError)
			}
			var cfg router.Config
			cfgPath := ""
			if path != "" {
				cfgPath = filepath.Join(path, "stucco")
			}
			if err := utils.LoadConfigFile(cfgPath, &cfg); err != nil {
				exitErr(err, initError)
			}
			var r project.Runtime
			switch runtime {
			case "stucco-js":
				r = runtimes.StuccoJS{}
			default:
				exitErr(errors.Errorf("runtime %s is not a valid value", runtime), initError)
			}
			p := project.Project{
				Config:             cfg,
				Output:             output,
				Overwrite:          overwrite,
				Path:               path,
				Runtime:            r,
				WriteLocalSettings: localSettings,
			}
			if err := p.Write(); err != nil {
				exitErr(err, initError)
			}
		},
	}
)

func init() {
	azureCmd.AddCommand(azureInitCmd)

	azureInitCmd.Flags().BoolVarP(&localSettings, "local-settings", "l", true, "Create local.settings.json")
	azureInitCmd.Flags().BoolVar(&overwrite, "overwrite", true, "Overwrite existing files")
	azureInitCmd.Flags().StringVarP(&path, "path", "p", wd, "Project path")
	azureInitCmd.Flags().StringVarP(&output, "output", "o", wd, "Output root path")
	azureInitCmd.Flags().StringVarP(&runtime, "runtime", "r", "", "Stucco runtime name")
}
