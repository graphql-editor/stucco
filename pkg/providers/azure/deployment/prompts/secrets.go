package prompts

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/printer"
	"github.com/graphql-editor/stucco/pkg/providers/azure/deployment"

	prompt "github.com/c-bata/go-prompt"
)

// SecretsSource queries user for secrets
type SecretsSource struct{}

// Select provides queries user for location of a resource
func (s SecretsSource) Select(ctx context.Context, context deployment.Context) (secrets map[string]string, err error) {
	if err = ctx.Err(); err != nil {
		return
	}
	secrets = make(map[string]string)
	printer.ColorPrintf("Please type your secrets, enter empty key to continue.\n")
	key := "start"
	for err == nil && key != "" {
		if err = ctx.Err(); err == nil {
			d := New(func() []prompt.Suggest {
				return []prompt.Suggest{}
			})
			key, err = d.Prompt(ctx, "Key: ")
			if err == nil && key != "" {
				var value string
				value, err = d.Prompt(ctx, "Value: ")
				secrets[key] = value
			}
		}
	}
	return
}
