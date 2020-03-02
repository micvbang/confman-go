package cli

import (
	"context"
	"fmt"
	"os"

	"text/tabwriter"

	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ListCommandInput struct {
	ServiceName string
	Key         string
	Format      string
	Reveal      bool
	Quiet       bool
}

func ConfigureListCommand(ctx context.Context, app *kingpin.Application, log confman.Logger, storage storage.Storage) {
	input := ListCommandInput{}

	cmd := app.Command("list", "Lists configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServiceName)

	cmd.Flag("reveal", "Reveal values").
		Envar("CONFMAN_REVEAL_VALUES").
		BoolVar(&input.Reveal)

	addFlagOutputFormat(cmd, &input.Format)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ListCommand(ctx, app, input, log, storage), "list")
		return nil
	})
}

func ListCommand(ctx context.Context, app *kingpin.Application, input ListCommandInput, log confman.Logger, storage storage.Storage) error {
	cm := confman.New(log, storage, input.ServiceName)
	config, err := cm.ReadAll(ctx)
	if err != nil {
		return err
	}

	if !input.Reveal {
		for key := range config {
			config[key] = "***"
		}
	}

	w := os.Stdout

	if input.Format == formatJSON {
		return outputJSON(w, cm.ServiceName(), config)
	}

	tw := tabwriter.NewWriter(w, 25, 4, 2, ' ', 0)

	fmt.Fprintf(tw, "Config for '%s'\n", cm.ServiceName())
	fmt.Fprintln(tw, "Key\tValue\t")
	fmt.Fprintln(tw, "=====\t=======\t")

	for key, value := range config {
		fmt.Fprintf(tw, "%s\t%s\t\n", key, value)
	}

	return tw.Flush()
}
