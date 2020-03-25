package prompts

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/printer"
	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment/config"

	prompt "github.com/c-bata/go-prompt"
	"github.com/pkg/errors"
)

// SubscriptionSource lists all available subscriptions and prompts user to pick one
type SubscriptionSource struct{}

// Select subscription queries user to choose a subscription if there's more than one available
func (sq SubscriptionSource) Select(ctx context.Context, subs []config.Subscription) (s config.Subscription, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	if len(subs) == 0 {
		err = errors.Errorf("no subscription found")
		return
	}
	s = subs[0]
	if len(subs) > 1 {
		choices := append([]config.Subscription{}, subs...)
		d := New(func() []prompt.Suggest {
			var suggestions []prompt.Suggest
			for _, sub := range choices {
				suggestions = append(suggestions, prompt.Suggest{Text: sub.ID, Description: fmt.Sprintf("Subscription named %s", sub.Name)})
			}
			choices = []config.Subscription{}
			return suggestions
		})
		printer.ColorPrintf("Please select your Azure subscription. By default %s will be chosen.\n\n", s.Name)
		var input string
		input, err = d.Prompt(ctx, "Subcription: ")
		if err == nil {
			for _, sub := range subs {
				if sub.ID == input {
					s = sub
				}
			}
		}
	}
	return
}
