package prompts

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/printer"
	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment"

	"github.com/AlekSi/pointer"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-06-01/subscriptions"
	prompt "github.com/c-bata/go-prompt"
	"github.com/pkg/errors"
)

// LocationSource is a paginated location querier
type LocationSource struct{}

// Select provides queries user for location of a resource
func (l LocationSource) Select(ctx context.Context, locations []subscriptions.Location) (location deployment.Location, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	location = deployment.Location("eastus")
	choices := append([]subscriptions.Location{}, locations...)
	d := New(func() []prompt.Suggest {
		suggestions := make([]prompt.Suggest, 0, len(locations))
		for _, loc := range choices {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        pointer.GetString(loc.Name),
				Description: pointer.GetString(loc.DisplayName),
			})
		}
		choices = nil
		return suggestions
	})
	printer.ColorPrintf("Please pick one of provided locations. By default %s will be chosen.\n\n", location)
	input, err := d.Prompt(ctx, "Location: ")
	if err == nil && input != "" {
		for _, loc := range locations {
			if pointer.GetString(loc.Name) == input {
				location = deployment.Location(pointer.GetString(loc.Name))
				break
			}
		}
		if input != string(location) {
			err = errors.Errorf("invalid location %s", input)
		}
	}
	return
}
