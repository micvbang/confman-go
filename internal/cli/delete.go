package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/micvbang/go-helpy/mapy"
	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/logger"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type DeleteCommandInput struct {
	ServiceName string
	Keys        []string
	Format      string
	Quiet       bool
	DeleteAll   bool
}

func ConfigureDeleteCommand(ctx context.Context, app *kingpin.Application, log logger.Logger, storage storage.Storage) {
	input := DeleteCommandInput{}

	cmd := app.Command("delete", "Deletes configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServiceName)

	cmd.Arg("keys", "Keys to delete").
		StringsVar(&input.Keys)

	cmd.Flag("delete-all-keys", "Ignore 'keys' argument and delete all keys for service").
		BoolVar(&input.DeleteAll)

	addFlagOutputFormat(cmd, &input.Format)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(DeleteCommand(ctx, app, input, log, storage), "list")
		return nil
	})
}

func DeleteCommand(ctx context.Context, app *kingpin.Application, input DeleteCommandInput, log logger.Logger, storage storage.Storage) error {
	cm := confman.New(log, storage, input.ServiceName)

	w := os.Stdout

	if !input.DeleteAll {
		err := cm.DeleteKeys(ctx, input.Keys)
		if err != nil {
			return err
		}
	} else {
		config, err := cm.ReadAll(ctx)
		if err != nil {
			return err
		}

		err = cm.DeleteAll(ctx)
		if err != nil {
			return err
		}

		input.Keys, _ = mapy.StringKeys(config)
	}

	if input.Format == formatJSON {
		return outputJSON(w, map[string]interface{}{
			cm.ServiceName(): input.Keys,
		})
	}

	for _, key := range input.Keys {
		fmt.Fprintf(w, "Deleted %s\n", cm.FormatKeyPath(key))
	}

	return nil
}
