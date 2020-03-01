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
	Quiet       bool
}

func ConfigureListCommand(ctx context.Context, app *kingpin.Application, log confman.Logger, storage storage.Storage) {
	input := ListCommandInput{}

	cmd := app.Command("list", "Lists configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServiceName)

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

	w := tabwriter.NewWriter(os.Stdout, 25, 4, 2, ' ', 0)

	fmt.Fprintf(w, "Config for '%s'\n", cm.ServiceName())
	fmt.Fprintln(w, "Key\tValue\t")
	fmt.Fprintln(w, "=====\t=======\t")

	for key, value := range config {
		fmt.Fprintf(w, "%s\t%s\t\n", key, value)
	}

	return w.Flush()
}
