package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/micvbang/confman-go/pkg/confman"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ReadCommandInput struct {
	ServicePath string
	Keys        []string
	Quiet       bool
	Format      string
}

func ConfigureReadCommand(ctx context.Context, app *kingpin.Application, log logger.Logger) {
	input := ReadCommandInput{}

	cmd := app.Command("read", "Reads a configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServicePath)

	cmd.Arg("keys", "Name of the keys to read").
		Required().
		StringsVar(&input.Keys)

	cmd.Flag("quiet", "Print only the value (only works for a single key").
		Short('q').
		Default("false").
		BoolVar(&input.Quiet)

	addFlagOutputFormat(cmd, &input.Format)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ReadCommand(ctx, input, os.Stdout, log, GlobalFlags.Storage), "read")
		return nil
	})
}

func ReadCommand(ctx context.Context, input ReadCommandInput, w io.Writer, log logger.Logger, storage storage.Storage) error {
	cm := confman.New(log, storage, input.ServicePath)
	config, err := cm.ReadKeys(ctx, input.Keys)
	if err != nil {
		return err
	}

	if input.Quiet && len(config) == 1 {
		for _, value := range config {
			fmt.Fprint(w, value)
		}

		return nil
	}

	if input.Format != formatText {
		return outputFormat(input.Format, w, map[string]interface{}{
			cm.ServicePath(): config,
		})
	}

	for key, value := range config {
		fmt.Fprintf(w, "%s = %s\n", cm.FormatKeyPath(key), value)
	}
	return nil
}
