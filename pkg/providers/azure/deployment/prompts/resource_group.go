package prompts

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/graphql-editor/stucco/pkg/printer"
	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment"

	"github.com/AlekSi/pointer"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	prompt "github.com/c-bata/go-prompt"
	"github.com/docker/docker/pkg/namesgenerator"
)

// ResourceGroupSource is a paginated prompt querier
type ResourceGroupSource struct{}

func randName() (name string, err error) {
	var suffixBytes [6]byte
	if _, err = rand.Read(suffixBytes[:]); err != nil {
		return
	}
	name += strings.Replace(namesgenerator.GetRandomName(0), "_", "-", -1)
	name += "-" + strings.ToLower(hex.EncodeToString(suffixBytes[:]))
	return
}

// Select provides an option to user to either create new resource group or select a new one from existing ones.
func (r ResourceGroupSource) Select(ctx context.Context, groupsClient resources.GroupsClient) (rg deployment.ResourceGroup, err error) {
	var resourceGroups []resources.Group
	groupListPage, err := groupsClient.List(ctx, "", pointer.ToInt32(6))
	if err != nil {
		return
	}
	randName, err := randName()
	if err != nil {
		return
	}
	rg = deployment.ResourceGroup{
		Name: randName,
	}
	d := New(func() []prompt.Suggest {
		var newRg []resources.Group
		if err != nil {
			return []prompt.Suggest{}
		}
		if groupListPage.NotDone() {
			newRg = groupListPage.Values()
			err = groupListPage.NextWithContext(ctx)
		}
		resourceGroups = append(resourceGroups, newRg...)
		var suggestions []prompt.Suggest
		for _, rg := range newRg {
			suggestions = append(suggestions, prompt.Suggest{
				Text: pointer.GetString(rg.Name),
				Description: fmt.Sprintf(
					"Resource group in region %s", pointer.GetString(rg.Location),
				),
			})
		}
		return suggestions
	})
	printer.ColorPrintf("Please type the name of your resource group or select an existing one. By default new one will be created with name %s.\n\n", rg.Name)
	input, err := d.Prompt(ctx, "Resource group name: ")
	if err == nil && input != "" {
		for _, arg := range resourceGroups {
			if pointer.GetString(arg.Name) == input {
				if rg.Name == input {
					if rg.Location != "" {
						// There are multiple resource groups with the same name
						// with the same name, remove location to query for it later.
						rg.Location = ""
					}
				} else {
					rg = deployment.ResourceGroup{
						Name:     pointer.GetString(arg.Name),
						Location: deployment.Location(pointer.GetString(arg.Location)),
					}
				}
			}
		}
		if rg.Name != input {
			rg = deployment.ResourceGroup{
				Name: input,
			}
		}
	}
	return
}
