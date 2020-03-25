package prompts

import (
	"context"
	"errors"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

type parser struct {
	prompt.ConsoleParser
	done bool
}

var (
	// ErrPromptInterrupt error is returned if Ctrl+C or Ctrl+D was typed during prompt
	ErrPromptInterrupt = struct{ error }{errors.New("prompt recieved interrupt")}
)

func (p *parser) Read() (b []byte, err error) {
	if p.done {
		return []byte{'\n'}, nil
	}
	b, err = p.ConsoleParser.Read()
	if err == nil {
		for _, c := range b {
			switch c {
			case byte(prompt.ControlC), byte(prompt.ControlD):
				p.done = true
				return []byte("^C"), nil
			}
		}
	}
	return
}

// NewSuggestions function used by DynamicPrompt to build suggestion list
type NewSuggestions func() []prompt.Suggest

// DynamicPrompt allows building paginated prompts with dynamic page fetching
type DynamicPrompt struct {
	NewSuggestions NewSuggestions
	completerState *prompt.CompletionManager
	suggestions    []prompt.Suggest
}

func tryMore(suggestions []prompt.Suggest, state *prompt.CompletionManager) bool {
	state.Next()
	_, ok := state.GetSelectedSuggestion()
	if !ok {
		state.Previous()
	}
	state.Previous()
	return !ok
}

func (d *DynamicPrompt) complete(in prompt.Document) []prompt.Suggest {
	suggestions := d.completerState.GetSuggestions()
	var noUpdate bool
	for tryMore(suggestions, d.completerState) && !noUpdate {
		newSuggestions := append(d.suggestions, d.NewSuggestions()...)
		// check if number of suggestions from source has been exhausted
		noUpdate = len(d.suggestions) == len(newSuggestions)
		d.suggestions = newSuggestions
		d.completerState.Update(in)
		suggestions = d.completerState.GetSuggestions()
	}
	return suggestions
}

func (d *DynamicPrompt) currentSuggestions(in prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(d.suggestions, in.GetWordBeforeCursor(), true)
}

func (d *DynamicPrompt) next() {
	d.completerState.Next()
}

func (d *DynamicPrompt) prev() {
	d.completerState.Previous()
}

func clean(in string, breakline bool) bool {
	return strings.HasSuffix(in, "^C")
}

// Prompt creates new input prompt with dynamic data
func (d *DynamicPrompt) Prompt(ctx context.Context, prefix string) (string, error) {
	p := prompt.New(
		nil,
		d.complete,
		prompt.OptionPrefix(prefix),
		prompt.OptionCompletionOnDown(),
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionPrefixTextColor(prompt.DefaultColor),
		prompt.OptionParser(&parser{
			ConsoleParser: prompt.NewStandardInputParser(),
		}),
		prompt.OptionAddKeyBind(
			prompt.KeyBind{
				Key: prompt.Up,
				Fn: func(*prompt.Buffer) {
					d.prev()
				},
			},
			prompt.KeyBind{
				Key: prompt.Down,
				Fn: func(*prompt.Buffer) {
					d.next()
				},
			},
			prompt.KeyBind{
				Key: prompt.Tab,
				Fn: func(*prompt.Buffer) {
					d.next()
				},
			},
		),
	)
	in := p.Input()
	if clean(in, false) {
		return "", ErrPromptInterrupt
	}
	return in, nil
}

// New creates new dynamic prompt with dynamic suggestions
func New(suggestions NewSuggestions) *DynamicPrompt {
	d := new(DynamicPrompt)
	d.completerState = prompt.NewCompletionManager(d.currentSuggestions, 6)
	d.NewSuggestions = suggestions
	return d
}
