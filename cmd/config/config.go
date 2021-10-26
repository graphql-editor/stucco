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
package configcmd

import (
	"os"
	"path/filepath"

	"github.com/graphql-editor/stucco/pkg/server"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func addToConfig() (*server.Config, error) {
	wd, err := os.Getwd()

	if err != nil {
		return nil, err
	}
	var cfg server.Config
	cfgPath := ""
	if wd != "" {
		cfgPath = filepath.Join(wd, "stucco")
	}
	if err := utils.LoadConfigFile(cfgPath, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// NewConfigCommand create new config command
func NewConfigCommand() *cobra.Command {
	configCommand := &cobra.Command{
		Use:   "config",
		Short: "basic stucco config",
	}
	configCommand.AddCommand(addCommand())
	return configCommand
}

func exitErr(err error, msg string) {
	klog.Fatalln(errors.Wrap(err, msg).Error())
}
