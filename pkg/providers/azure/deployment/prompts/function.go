package prompts

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/graphql-editor/stucco/pkg/printer"
	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment"

	"github.com/AlekSi/pointer"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-06-01/subscriptions"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2019-08-01/web"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	prompt "github.com/c-bata/go-prompt"
	"github.com/pkg/errors"
)

// FunctionAppSource is a paginated prompt source for function data
type FunctionAppSource struct{}

func checkFunctionAppPlanSourceError(err error) error {
	if err == nil {
		return err
	}
	detailedError, ok := err.(autorest.DetailedError)
	if !ok || detailedError.StatusCode != http.StatusNotFound {
		return err
	}
	azureError, ok := detailedError.Original.(*azure.RequestError)
	if !ok || azureError.ServiceError == nil || azureError.ServiceError.Code != "ResourceGroupNotFound" {
		return err
	}
	return nil
}

var azureFunctionsLocations = func() (locations []subscriptions.Location) {
	for _, l := range deployment.FunctionLocations {
		loc := l
		locations = append(locations, subscriptions.Location{
			Name:        &loc.Slug,
			DisplayName: &loc.Name,
		})
	}
	return
}()

func queryLocation(ctx context.Context, fp deployment.FunctionApp) (loc deployment.Location, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	if deployment.ValidFunctionAppLocation(fp) != "" {
		return fp.Location, nil
	}
	if fp.Location != "" {
		printer.ColorPrintf("'%s' is not a valid function app location. ", string(fp.Location))
	}
	var source LocationSource
	return source.Select(ctx, azureFunctionsLocations)
}

func querySku(ctx context.Context, plan deployment.FunctionAppPlan) (sku string, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	if deployment.ValidFunctionAppSku(plan) != "" {
		return plan.Sku, nil
	}
	if plan.Sku != "" {
		printer.ColorPrintf("'%s' is not a valid plan sku. ", plan.Sku)
	}
	if err = ctx.Err(); err != nil {
		return
	}
	sku = "B1"
	choices := append([]deployment.Plan{}, deployment.SupportedPlans...)
	d := New(func() []prompt.Suggest {
		suggestions := make([]prompt.Suggest, 0, len(choices))
		for _, plan := range choices {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        plan.Sku,
				Description: plan.Display,
			})
		}
		choices = nil
		return suggestions
	})
	printer.ColorPrintf("Please pick one of provided SKUs. By default %s will be chosen.\n\n", sku)
	input, err := d.Prompt(ctx, "SKU: ")
	if err == nil && input != "" {
		for _, supportedPlan := range deployment.SupportedPlans {
			if supportedPlan.Sku == input {
				sku = input
				break
			}
		}
		if input != string(sku) {
			err = errors.Errorf("invalid sku %s", input)
		}
	}
	return
}

func queryPlan(ctx context.Context, context deployment.Context, fp deployment.FunctionApp, client web.BaseClient) (plan deployment.FunctionAppPlan, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	var appPlans []web.AppServicePlan
	plansClient := web.AppServicePlansClient{
		BaseClient: client,
	}
	plans, err := plansClient.ListByResourceGroup(
		ctx,
		context.ResourceGroup.Name,
	)
	if err = checkFunctionAppPlanSourceError(err); err != nil {
		return
	}
	plan.Name = fp.Name
	d := New(func() []prompt.Suggest {
		var suggestions []prompt.Suggest
		for plans.NotDone() && err == nil {
			newPlans := plans.Values()
			err = plans.NextWithContext(ctx)
			appPlans = append(appPlans, newPlans...)
			for _, plan := range newPlans {
				if pointer.GetString(plan.Location) == string(fp.Location) {
					suggestions = append(suggestions, prompt.Suggest{
						Text: pointer.GetString(plan.Name),
						Description: fmt.Sprintf(
							"%s tier plan", pointer.GetString(plan.Sku.Name),
						),
					})
				}
			}
			if len(newPlans) <= len(suggestions) {
				break
			}
		}
		return suggestions
	})
	printer.ColorPrintf("Please type the name of your FunctionApp plan or select an existing one. By default new one will be created with name %s.\n\n", fp.Name)
	input, err := d.Prompt(ctx, "FunctionApp plan: ")
	if err == nil {
		if input != "" {
			for _, arg := range appPlans {
				if pointer.GetString(arg.Name) == input {
					plan = deployment.FunctionAppPlan{
						Name: pointer.GetString(arg.Name),
						Sku:  pointer.GetString(arg.Sku.Name),
					}
				}
			}
		}
		if plan.Sku == "" {
			plan.Sku, err = querySku(ctx, plan)
		}
	}
	return
}

func queryWorkers(ctx context.Context) (int32, error) {
	d := New(func() []prompt.Suggest {
		return []prompt.Suggest{}
	})
	printer.ColorPrintf("Please type the number of workers you would want to have in your app. Defaults to 1.\n\n")
	input, err := d.Prompt(ctx, "Number of workers: ")
	i64 := int64(1)
	if err == nil && input != "" {
		i64, err = strconv.ParseInt(input, 10, 32)
	}
	return int32(i64), err
}

func queryImage(ctx context.Context) (string, error) {
	d := New(func() []prompt.Suggest {
		return []prompt.Suggest{}
	})
	printer.ColorPrintf("Please type the name of your FunctionApp docker image.\n\n")
	input, err := d.Prompt(ctx, "FunctionApp image: ")
	return input, err
}

func getAPINameFromSite(site web.Site) (apiName string, ok bool) {
	if len(site.Tags) > 0 && pointer.GetString(site.Tags["managedBy"]) == "stucco" {
		apiName = pointer.GetString(site.Tags["apiName"])
	}
	ok = apiName != ""
	return
}

func getSiteSuggestion(site web.Site) (s prompt.Suggest) {
	s.Description = fmt.Sprintf("Function app in %s", pointer.GetString(site.Location))
	if apiName, ok := getAPINameFromSite(site); ok {
		s.Text = apiName
	} else {
		s.Text = pointer.GetString(site.Name)
	}
	return
}

// Select provides an option to user to either create new resource group or select a new one from existing ones.
func (r FunctionAppSource) Select(ctx context.Context, context deployment.Context, client web.BaseClient) (fp deployment.FunctionApp, err error) {
	var sites []web.Site
	appsClient := web.AppsClient{
		BaseClient: client,
	}
	sitesList, err := appsClient.ListByResourceGroup(
		ctx,
		context.ResourceGroup.Name,
		pointer.ToBool(false),
	)
	fp = deployment.FunctionApp{
		Name:     context.ResourceGroup.Name,
		Location: context.ResourceGroup.Location,
	}
	unique := make(map[prompt.Suggest]struct{})
	d := New(func() []prompt.Suggest {
		var suggestions []prompt.Suggest
		for sitesList.NotDone() && err == nil {
			newSites := sitesList.Values()
			err = sitesList.NextWithContext(ctx)
			sites = append(sites, newSites...)
			for _, site := range newSites {
				if loc := deployment.ValidFunctionAppLocation(deployment.FunctionApp{
					Location: deployment.Location(pointer.GetString(site.Location)),
				}); loc != "" {
					suggestion := getSiteSuggestion(site)
					if _, ok := unique[suggestion]; !ok {
						unique[suggestion] = struct{}{}
						suggestions = append(suggestions, suggestion)
					}
				}
			}
			if len(newSites) <= len(suggestions) {
				break
			}
		}
		return suggestions
	})
	printer.ColorPrintf("Please type the name of your FunctionApp or select an existing one. By default new one will be created with name %s.\n\n", fp.Name)
	input, err := d.Prompt(ctx, "FunctionApp: ")
	if err == nil && input != "" {
		for _, arg := range sites {
			if pointer.GetString(arg.Name) == input {
				fp = deployment.FunctionApp{
					Name:     pointer.GetString(arg.Name),
					Location: deployment.Location(pointer.GetString(arg.Location)),
				}
			}
		}
		fp.Location, err = queryLocation(ctx, fp)
	}
	if err == nil {
		fp.Plan, err = queryPlan(ctx, context, fp, client)
	}
	if err == nil {
		fp.Image.Repository, err = queryImage(ctx)
	}
	if err == nil {
		fp.Plan.Workers, err = queryWorkers(ctx)
	}
	return
}
