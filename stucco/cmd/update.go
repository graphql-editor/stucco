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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/graphql-editor/stucco/pkg/providers/azure/configs"
	"github.com/graphql-editor/stucco/pkg/providers/azure/project"
	"github.com/graphql-editor/stucco/pkg/providers/azure/project/runtimes"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

const (
	updateError = "could not update project"
)

// azureUpdateCmd represents the update command
var azureUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update azure functions with new configuration",
	Run: func(cmd *cobra.Command, args []string) {
		hostJSON := filepath.Join(output, "host.json")
		_, err := os.Stat(hostJSON)
		if err != nil {
			if os.IsNotExist(err) {
				err = errors.Errorf("host.json does not exist, was project initialized?")
			}
			exitErr(err, updateError)
		}
		var cfg router.Config
		cfgPath := ""
		if path != "" {
			cfgPath = filepath.Join(path, "stucco")
		}
		if err := utils.LoadConfigFile(cfgPath, &cfg); err != nil {
			exitErr(err, updateError)
		}
		if runtime == "" {
			localSettingsJSON := filepath.Join(output, "local.settings.json")
			if b, err := ioutil.ReadFile(localSettingsJSON); err == nil {
				var localSettings configs.LocalSettings
				if err := json.Unmarshal(b, &localSettings); err == nil {
					if rt, ok := localSettings.Values["FUNCTIONS_WORKER_RUNTIME"]; ok {
						runtime = rt
					}
				} else {
					klog.Warningf("%s could not be read: %v", localSettingsJSON, err)
				}
			} else {
				klog.Warningf("%s could not be read: %v", localSettingsJSON, err)
			}
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
		if err := p.Update(); err != nil {
			exitErr(err, initError)
		}
	},
}

func init() {
	azureCmd.AddCommand(azureUpdateCmd)

	azureUpdateCmd.Flags().BoolVarP(&localSettings, "local-settings", "l", true, "Create local.settings.json")
	azureUpdateCmd.Flags().BoolVar(&overwrite, "overwrite", true, "Overwrite existing files")
	azureUpdateCmd.Flags().StringVarP(&path, "path", "p", wd, "Project path")
	azureUpdateCmd.Flags().StringVarP(&output, "output", "o", wd, "Output root path")
	azureUpdateCmd.Flags().StringVarP(&runtime, "runtime", "r", "", "Stucco runtime name")
}
