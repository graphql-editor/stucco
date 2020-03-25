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
	"github.com/spf13/cobra"
)

// NewAzureCommand represents the azure command
func NewAzureCommand() *cobra.Command {
	var azureCmd = &cobra.Command{
		Use:   "azure",
		Short: "Set of tools to help with Azure Functions and stucco integration",
	}
	azureCmd.AddCommand(NewInitCommand())
	azureCmd.AddCommand(NewUpdateCommand())
	azureCmd.AddCommand(NewDeployCommand())
	azureCmd.AddCommand(NewStartCommand())
	return azureCmd
}
