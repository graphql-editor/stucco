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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/graphql-editor/stucco/pkg/providers/azure/configs"
	"github.com/graphql-editor/stucco/pkg/providers/azure/project"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

const (
	updateError = "could not update project"
)

// NewUpdateCommand for azure
func NewUpdateCommand() *cobra.Command {
	genOpts := newGenOpts()
	var updateCommand = &cobra.Command{
		Use:   "update",
		Short: "Update azure functions with new configuration",
		Run: func(cmd *cobra.Command, args []string) {
			hostJSON := filepath.Join(genOpts.output, "host.json")
			_, err := os.Stat(hostJSON)
			if err != nil {
				if os.IsNotExist(err) {
					err = errors.Errorf("host.json does not exist, was project initialized?")
				}
				exitErr(err, updateError)
			}
			var cfg router.Config
			cfgPath := ""
			if genOpts.path != "" {
				cfgPath = filepath.Join(genOpts.path, "stucco")
			}
			if err := utils.LoadConfigFile(cfgPath, &cfg); err != nil {
				exitErr(err, updateError)
			}
			if genOpts.runtime == "" {
				localSettingsJSON := filepath.Join(genOpts.output, "local.settings.json")
				if b, err := ioutil.ReadFile(localSettingsJSON); err == nil {
					var localSettings configs.LocalSettings
					if err := json.Unmarshal(b, &localSettings); err == nil {
						if rt, ok := localSettings.Values["FUNCTIONS_WORKER_RUNTIME"]; ok {
							genOpts.runtime = rt
						}
					} else {
						klog.Warningf("%s could not be read: %v", localSettingsJSON, err)
					}
				} else {
					klog.Warningf("%s could not be read: %v", localSettingsJSON, err)
				}
			}
			if err := genOpts.validate(); err != nil {
				exitErr(err, updateError)
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
			if err := p.Update(); err != nil {
				exitErr(err, updateError)
			}
		},
	}

	genOpts.setupCommand(updateCommand)
	return updateCommand
}
