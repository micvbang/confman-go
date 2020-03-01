package cli

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ReadCommandInput struct {
	ServiceName string
	Key         string
	Quiet       bool
}

func ConfigureReadCommand(ctx context.Context, app *kingpin.Application, log confman.Logger, storage storage.Storage) {
	input := ReadCommandInput{}

	cmd := app.Command("read", "Reads a configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServiceName)

	cmd.Arg("key", "Name of the key").
		Required().
		StringVar(&input.Key)

	cmd.Flag("quiet", "Print only the value").
		Short('q').
		Default("false").
		BoolVar(&input.Quiet)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ReadCommand(ctx, app, input, log, storage), "read")
		return nil
	})
}

func ReadCommand(ctx context.Context, app *kingpin.Application, input ReadCommandInput, log confman.Logger, storage storage.Storage) error {
	cm := confman.New(log, storage, input.ServiceName)
	value, err := cm.Read(ctx, input.Key)
	if err != nil {
		return err
	}

	w := os.Stdout

	if input.Quiet {
		fmt.Fprint(w, value)
		return nil
	}

	fmt.Fprintf(w, "%s: %s", cm.FormatKeyPath(input.Key), value)
	return nil
}
