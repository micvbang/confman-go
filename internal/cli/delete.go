package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/micvbang/go-helpy/stringy"
	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type DeleteCommandInput struct {
	ServiceName string
	Keys        []string
	Quiet       bool
}

const deleteAllKeysAlias = "*"

func ConfigureDeleteCommand(ctx context.Context, app *kingpin.Application, log confman.Logger, storage storage.Storage) {
	input := DeleteCommandInput{}

	cmd := app.Command("delete", "Deletes configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServiceName)

	cmd.Arg("keys", fmt.Sprintf("Keys to delete. Use '%s' for all keys", deleteAllKeysAlias)).
		StringsVar(&input.Keys)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(DeleteCommand(ctx, app, input, log, storage), "list")
		return nil
	})
}

func DeleteCommand(ctx context.Context, app *kingpin.Application, input DeleteCommandInput, log confman.Logger, storage storage.Storage) error {
	cm := confman.New(log, storage, input.ServiceName)

	lookup := stringy.ToSet(input.Keys)
	allKeys := lookup.Contains(deleteAllKeysAlias)

	w := os.Stdout

	if allKeys {
		err := cm.DeleteAll(ctx)
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "Deleted %s", cm.FormatKeyPath(deleteAllKeysAlias))
		return nil
	}

	err := cm.DeleteKeys(ctx, input.Keys)
	if err != nil {
		return err
	}

	for _, key := range input.Keys {
		fmt.Fprintf(w, "Deleted %s\n", cm.FormatKeyPath(key))
	}

	return nil
}
