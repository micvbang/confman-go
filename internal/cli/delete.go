package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/micvbang/confman-go/pkg/confman"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/micvbang/go-helpy/mapy"
	"gopkg.in/alecthomas/kingpin.v2"
)

type DeleteCommandInput struct {
	ServicePath string
	Keys        []string
	Format      string
	Quiet       bool
	DeleteAll   bool
}

func ConfigureDeleteCommand(ctx context.Context, app *kingpin.Application, log logger.Logger) {
	input := DeleteCommandInput{}

	cmd := app.Command("delete", "Deletes configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServicePath)

	cmd.Arg("keys", "Keys to delete").
		StringsVar(&input.Keys)

	cmd.Flag("delete-all-keys", "Ignore 'keys' argument and delete all keys for service").
		BoolVar(&input.DeleteAll)

	addFlagOutputFormat(cmd, &input.Format)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(DeleteCommand(ctx, input, os.Stdout, log, GlobalFlags.Storage), "delete")
		return nil
	})
}

func DeleteCommand(ctx context.Context, input DeleteCommandInput, w io.Writer, log logger.Logger, storage storage.Storage) error {
	cm := confman.New(log, storage, input.ServicePath)

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

		input.Keys = mapy.Keys(config)
	}

	if input.Format != formatText {
		return outputFormat(input.Format, w, map[string]interface{}{
			cm.ServicePath(): input.Keys,
		})
	}

	for _, key := range input.Keys {
		fmt.Fprintf(w, "Deleted %s\n", cm.FormatKeyPath(key))
	}

	return nil
}
