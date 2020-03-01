package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/99designs/aws-vault/prompt"

	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type AddCommandInput struct {
	ServiceName string
	Key         string
	Value       string
}

func ConfigureAddCommand(ctx context.Context, app *kingpin.Application, log confman.Logger, storage storage.Storage) {
	input := AddCommandInput{}

	cmd := app.Command("add", "Adds a configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServiceName)

	cmd.Arg("key", "Name of the key").
		Required().
		StringVar(&input.Key)

	cmd.Flag("value", "Value to add (don't use this flag for secret values)").
		Short('v').
		StringVar(&input.Value)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(AddCommand(ctx, app, input, log, storage), "add")
		return nil
	})
}

func AddCommand(ctx context.Context, app *kingpin.Application, input AddCommandInput, log confman.Logger, storage storage.Storage) error {
	cm := confman.New(log, storage, input.ServiceName)

	var err error
	value := input.Value
	if len(value) == 0 {
		value, err = prompt.TerminalPrompt("Enter value: ")
		if err != nil {
			return err
		}
	}

	err = cm.Add(ctx, input.Key, value)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s = '%s'", cm.FormatKeyPath(input.Key), value)
	return nil
}