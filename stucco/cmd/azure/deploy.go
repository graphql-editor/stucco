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
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-06-01/subscriptions"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2019-08-01/web"
	"github.com/graphql-editor/stucco/pkg/printer"
	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment"
	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment/config"
	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment/prompts"
	"github.com/spf13/cobra"
)

const (
	deployError = "could not deploy project"
)

// NewDeployCommand returns new deploy command
func NewDeployCommand() *cobra.Command {
	var (
		cliPath        string
		interactive    bool
		dryRun         bool
		projectConfig  string
		subscription   config.Subscription
		rg             deployment.ResourceGroup
		loc            deployment.Location
		storageAccount deployment.StorageAccount
		blobContainer  deployment.BlobContainer
		functionApp    deployment.FunctionApp
		secretsFile    string
		secrets        secrets
	)
	deployCommand := &cobra.Command{
		Use:   "deploy",
		Short: "EXPERIMENTAL: Deploy Azure Functions from stucco.json",
		Long:  `Deploys Azure Functions to Azure cloud`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			exitCh := make(chan os.Signal, 1)
			signal.Notify(exitCh, os.Interrupt)
			go func() {
				<-exitCh
				cancel()
				os.Exit(1)
			}()
			cfgopts := []config.Option{
				config.SubscriptionSourceOpt(subscriptionsSource{
					interactive:  interactive,
					subscription: subscription,
				}),
			}
			if cliPath != "" {
				cfgopts = append([]config.Option{config.CLIPathOpt(cliPath)}, cfgopts...)
			}
			cfg, err := config.NewConfig(
				ctx,
				cfgopts...,
			)
			if err != nil {
				printer.ErrorPrintf("Could not prepare deployment: %v", err)
				os.Exit(1)
			}
			if cfg.SubscriptionID != "" && !dryRun {
				printer.NotePrintf("Using subscription %s.\n\n", cfg.SubscriptionID)
			}
			var depCtx deployment.Context
			if projectConfig != "" {
				f, err := os.Open(projectConfig)
				if err != nil {
					printer.ErrorPrintf("Could not prepare deployment: %v", err)
					os.Exit(1)
				}
				err = json.NewDecoder(f).Decode(&depCtx)
				if ferr := f.Close(); ferr != nil {
					printer.ErrorPrintf("could not close file", ferr)
				}
				if err != nil {
					printer.ErrorPrintf("Could not prepare deployment: %v", err)
					os.Exit(1)
				}
			}
			opts := []deployment.Option{
				deployment.ResourceGroupSourceOpt(resourceGroupSource{
					interactive: interactive,
					ctx:         depCtx,
					rg:          rg,
				}),
				deployment.ResourceGroupLocationSourceOpt(locationSource{
					interactive: interactive,
					ctx:         depCtx,
					loc:         loc,
				}),
				deployment.StorageAccountSourceOpt(storageAccountSource{
					interactive:    interactive,
					ctx:            depCtx,
					storageAccount: storageAccount,
				}),
				deployment.BlobContainerSourceOpt(blobContainerSource{
					interactive:   interactive,
					ctx:           depCtx,
					blobContainer: blobContainer,
				}),
				deployment.FunctionAppOpt(functionAppSource{
					interactive: interactive,
					ctx:         depCtx,
					functionApp: functionApp,
				}),
				deployment.SecretsSourceOpt(secretsSource{
					interactive: interactive,
					secretsFile: secretsFile,
				}),
			}
			if schema, err := ioutil.ReadFile("schema.graphql"); err == nil {
				opts = append(opts, deployment.SchemaOpt(string(schema)))
			}
			if stuccoJSON, err := ioutil.ReadFile("stucco.json"); err == nil {
				opts = append(opts, deployment.StuccoJSONOpt(string(stuccoJSON)))
			}
			b := deployment.NewBuilder(
				cfg,
				opts...,
			)
			depCtx, err = deployment.BuildContext(ctx, b)
			if err != nil {
				printer.ErrorPrintf("Could not prepare deployment: %v", err)
				os.Exit(1)
			}
			ctxJSON, _ := json.MarshalIndent(depCtx, "", "\t")
			if interactive && !dryRun {
				printer.ColorPrintf("I'm going to deploy a function with this configuration: \n%s\nContinue [y/n]?: ", string(ctxJSON))
				reader := bufio.NewReader(os.Stdin)
				yn, err := reader.ReadByte()
				if err != nil || yn != 'y' {
					os.Exit(1)
				}
			}
			if dryRun {
				printer.ColorPrintf("%s", string(ctxJSON))
			} else {
				deployClient := deployment.DeployClient{
					Config: cfg,
				}
				if err := deployClient.Deploy(ctx, depCtx); err != nil {
					printer.ErrorPrintf("Could deploy: %v", err)
					os.Exit(1)
				}
			}
		},
	}
	deployCommand.Flags().StringVar(&cliPath, "azure-cli", "", "Optional Azure CLI path if not in default system PATH")
	deployCommand.Flags().BoolVarP(&interactive, "interactive", "i", false, "Enable interactive deployment")
	deployCommand.Flags().BoolVar(&dryRun, "dry-run", false, "Do a dry run. Print deployment config without doing any actual changes")
	deployCommand.Flags().StringVarP(&projectConfig, "config", "c", "", "Path to a deployment configuration")
	deployCommand.Flags().StringVar(&subscription.ID, "subscription-id", "", "Azure subscription ID")
	deployCommand.Flags().StringVar((*string)(&loc), "location", "", "Location of project in Azure Cloud")
	deployCommand.Flags().StringVar(&rg.Name, "resource-group-name", "", "Name of Azure Cloud resource group name")
	deployCommand.Flags().StringVar((*string)(&storageAccount), "storage-account", "", "Name of storage account with project data")
	deployCommand.Flags().StringVar((*string)(&blobContainer), "blob-container", "", "Name of blob container with project data")
	deployCommand.Flags().StringVar(&functionApp.Name, "function-app-name", "", "Function app name")
	deployCommand.Flags().StringVar(&functionApp.Plan.Name, "function-app-plan-name", "", "Function app plan name")
	deployCommand.Flags().StringVar(&functionApp.Plan.Sku, "function-app-plan-sku", "", "Function app plan sku (B1, EP1 etc.)")
	deployCommand.Flags().Int32Var(&functionApp.Plan.Workers, "function-app-plan-workers", 0, "Number of workers for function plan")
	deployCommand.Flags().StringVar((*string)(&functionApp.Location), "function-app-location", "", "Location of function app. Overrides location set by location option")
	deployCommand.Flags().StringVar(&functionApp.Image.Repository, "function-app-image", "", "Docker image with function image")
	deployCommand.Flags().StringVar(&functionApp.Image.Registry, "function-app-image-registry", "", "Docker registry with function image")
	deployCommand.Flags().StringVar(&functionApp.Image.Username, "function-app-image-username", "", "Username for docker registry with function image")
	deployCommand.Flags().StringVar(&functionApp.Image.Password, "function-app-image-passowrd", "", "Password for docker registry with function image")
	deployCommand.Flags().StringVar(&secretsFile, "secrets-file", "", "Path to JSON file with an object of key:value pairs with secrets")
	deployCommand.Flags().Var(&secrets, "secret", "App secret. Can be repeated multiple times")
	return deployCommand
}

type subscriptionsSource struct {
	interactive  bool
	subscription config.Subscription
}

func (s subscriptionsSource) Select(ctx context.Context, subs []config.Subscription) (config.Subscription, error) {
	sub := s.subscription
	if subID := os.Getenv("AZURE_SUBSCRIPTION_ID"); subID != "" {
		sub.ID = subID
	}
	if s.interactive {
		prompt := prompts.SubscriptionSource{}
		psub, err := prompt.Select(ctx, subs)
		if err != nil {
			return config.Subscription{}, err
		}
		sub = psub
	}
	return sub, nil
}

func mergeResourceGroup(dst *deployment.ResourceGroup, src deployment.ResourceGroup) {
	if src.Name != "" {
		dst.Name = src.Name
	}
}

type resourceGroupSource struct {
	interactive bool
	ctx         deployment.Context
	rg          deployment.ResourceGroup
}

func (r resourceGroupSource) Select(ctx context.Context, client resources.GroupsClient) (deployment.ResourceGroup, error) {
	rg := r.ctx.ResourceGroup
	mergeResourceGroup(&rg, r.rg)
	if rgName := os.Getenv("AZURE_RESOURCE_GROUP_NAME"); rgName != "" {
		rg.Name = rgName
	}
	if r.interactive {
		prompt := prompts.ResourceGroupSource{}
		prg, err := prompt.Select(ctx, client)
		if err != nil {
			return deployment.ResourceGroup{}, err
		}
		mergeResourceGroup(&rg, prg)
	}
	return rg, nil
}

func mergeLocation(dst *deployment.Location, src deployment.Location) {
	if src != "" {
		*dst = src
	}
}

type locationSource struct {
	interactive bool
	ctx         deployment.Context
	loc         deployment.Location
}

func (l locationSource) Select(ctx context.Context, locations []subscriptions.Location) (deployment.Location, error) {
	loc := l.ctx.ResourceGroup.Location
	mergeLocation(&loc, l.loc)
	if envLoc := os.Getenv("AZURE_PROJECT_LOCATION"); envLoc != "" {
		loc = deployment.Location(envLoc)
	}
	if l.interactive {
		prompt := prompts.LocationSource{}
		l, err := prompt.Select(ctx, locations)
		if err != nil {
			return "", err
		}
		mergeLocation(&loc, l)
	}
	return loc, nil
}

type storageAccountSource struct {
	interactive    bool
	ctx            deployment.Context
	storageAccount deployment.StorageAccount
}

func mergeStorageAccount(dst *deployment.StorageAccount, src deployment.StorageAccount) {
	if src != "" {
		*dst = src
	}
}

func (s storageAccountSource) Select(ctx context.Context, client storage.AccountsClient, context deployment.Context) (deployment.StorageAccount, error) {
	sa := s.ctx.StorageAccount
	mergeStorageAccount(&sa, s.storageAccount)
	if envSa := os.Getenv("AZURE_STORAGE_ACCOUNT"); envSa != "" {
		sa = deployment.StorageAccount(envSa)
	}
	if s.interactive {
		prompt := prompts.StorageAccountSource("")
		psa, err := prompt.Select(ctx, client, context)
		if err != nil {
			return "", err
		}
		mergeStorageAccount(&sa, psa)
	}
	return sa, nil
}

func mergeBlobContainer(dst *deployment.BlobContainer, src deployment.BlobContainer) {
	if src != "" {
		*dst = src
	}
}

type blobContainerSource struct {
	interactive   bool
	ctx           deployment.Context
	blobContainer deployment.BlobContainer
}

func (b blobContainerSource) Select(ctx context.Context, client storage.BlobContainersClient, context deployment.Context) (deployment.BlobContainer, error) {
	bc := b.ctx.BlobContainer
	mergeBlobContainer(&bc, b.blobContainer)
	if envBC := os.Getenv("AZURE_BLOB_CONTAINER"); envBC != "" {
		bc = deployment.BlobContainer(envBC)
	}
	if b.interactive {
		prompt := prompts.BlobContainerSource("")
		pbc, err := prompt.Select(ctx, client, context)
		if err != nil {
			return "", err
		}
		mergeBlobContainer(&bc, pbc)
	}
	return bc, nil
}

type functionAppSource struct {
	interactive bool
	ctx         deployment.Context
	functionApp deployment.FunctionApp
}

func mergeFunctionApp(dst *deployment.FunctionApp, src deployment.FunctionApp) {
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.Image.Repository != "" {
		dst.Image.Repository = src.Image.Repository
	}
	if src.Image.Registry != "" {
		dst.Image.Registry = src.Image.Registry
	}
	if src.Image.Username != "" {
		dst.Image.Username = src.Image.Username
	}
	if src.Image.Password != "" {
		dst.Image.Password = src.Image.Password
	}
	if src.Plan.Name != "" {
		dst.Plan.Name = src.Plan.Name
	}
	if src.Plan.Sku != "" {
		dst.Plan.Sku = src.Plan.Sku
	}
	if src.Plan.Workers != 0 {
		dst.Plan.Workers = src.Plan.Workers
	}
}

func (f functionAppSource) Select(ctx context.Context, context deployment.Context, client web.BaseClient) (deployment.FunctionApp, error) {
	fa := f.ctx.FunctionApp
	mergeFunctionApp(&fa, f.functionApp)
	if faName := os.Getenv("AZURE_FUNCTION_APP_NAME"); faName != "" {
		fa.Name = faName
	}
	if faImage := os.Getenv("AZURE_FUNCTION_APP_IMAGE"); faImage != "" {
		fa.Image.Repository = faImage
	}
	if faImageReg := os.Getenv("AZURE_FUNCTION_APP_IMAGE_REGISTRY"); faImageReg != "" {
		fa.Image.Registry = faImageReg
	}
	if faImageUser := os.Getenv("AZURE_FUNCTION_APP_IMAGE_REGISTRY_USERNAME"); faImageUser != "" {
		fa.Image.Username = faImageUser
	}
	if faImagePassword := os.Getenv("AZURE_FUNCTION_APP_IMAGE_REGISTRY_PASSWORD"); faImagePassword != "" {
		fa.Image.Password = faImagePassword
	}
	if faPlanName := os.Getenv("AZURE_FUNCTION_APP_PLAN_NAME"); faPlanName != "" {
		fa.Plan.Name = faPlanName
	}
	if faPlanSku := os.Getenv("AZURE_FUNCTION_APP_PLAN_SKU"); faPlanSku != "" {
		fa.Plan.Sku = faPlanSku
	}
	if faPlanWorkers := os.Getenv("AZURE_FUNCTION_APP_PLAN_WORKERS"); faPlanWorkers != "" {
		i, err := strconv.ParseInt(faPlanWorkers, 10, 32)
		if err != nil {
			return deployment.FunctionApp{}, err
		}
		fa.Plan.Workers = int32(i)
	}
	if f.interactive {
		prompt := prompts.FunctionAppSource{}
		pfa, err := prompt.Select(ctx, context, client)
		if err != nil {
			return deployment.FunctionApp{}, err
		}
		mergeFunctionApp(&fa, pfa)
	}
	if f.functionApp.Plan.Workers == 0 {
		f.functionApp.Plan.Workers = 1
	}
	return fa, nil
}

type secrets map[string]string

func (s *secrets) String() string {
	var secrets []string
	for k, v := range *s {
		secrets = append(secrets, k+"="+v)
	}
	return strings.Join(secrets, ",")
}
func (s *secrets) Set(v string) error {
	if *s == nil {
		*s = make(secrets)
	}
	parts := strings.Split(v, "=")
	if len(parts) != 2 {
		return errors.New("secret must be in form of <KEY>=<VALUE>")
	}
	(*s)[parts[0]] = parts[1]
	return nil
}
func (s *secrets) Type() string {
	return "secrets"
}

type secretsSource struct {
	interactive bool
	secretsFile string
	secretsVars secrets
}

func (s secretsSource) Select(ctx context.Context, context deployment.Context) (map[string]string, error) {
	secrets := make(secrets)
	if s.secretsFile != "" {
		f, err := os.Open(s.secretsFile)
		if err != nil {
			return nil, err
		}
		err = json.NewDecoder(f).Decode(&secrets)
		if ferr := f.Close(); ferr != nil {
			printer.ErrorPrintf("could not close file: ", ferr)
		}
		if err != nil {
			return nil, err
		}
	}
	for k, v := range s.secretsVars {
		secrets[k] = v
	}
	if s.interactive {
		prompt := prompts.SecretsSource{}
		sec, err := prompt.Select(ctx, context)
		if err != nil {
			return nil, err
		}
		for k, v := range sec {
			secrets[k] = v
		}
	}
	return (map[string]string)(secrets), nil
}
