package cli

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/logger"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ReadCommandInput struct {
	ServiceName string
	Keys        []string
	Quiet       bool
	Format      string
}

func ConfigureReadCommand(ctx context.Context, app *kingpin.Application, log logger.Logger) {
	input := ReadCommandInput{}

	cmd := app.Command("read", "Reads a configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServiceName)

	cmd.Arg("keys", "Name of the keys to read").
		Required().
		StringsVar(&input.Keys)

	cmd.Flag("quiet", "Print only the value (only works for a single key").
		Short('q').
		Default("false").
		BoolVar(&input.Quiet)

	addFlagOutputFormat(cmd, &input.Format)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ReadCommand(ctx, app, input, log, GlobalFlags.Storage), "read")
		return nil
	})
}

func ReadCommand(ctx context.Context, app *kingpin.Application, input ReadCommandInput, log logger.Logger, storage storage.Storage) error {
	cm := confman.New(log, storage, input.ServiceName)
	config, err := cm.ReadKeys(ctx, input.Keys)
	if err != nil {
		return err
	}

	w := os.Stdout

	if input.Quiet && len(config) == 1 {
		for _, value := range config {
			fmt.Fprint(w, value)
		}

		return nil
	}

	if input.Format == formatJSON {
		return outputJSON(w, map[string]interface{}{
			cm.ServiceName(): config,
		})
	}

	for key, value := range config {
		fmt.Fprintf(w, "%s=%s\n", cm.FormatKeyPath(key), value)
	}
	return nil
}
