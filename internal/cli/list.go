package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	configKeys, err := cm.ReadAllMetadata(ctx)
	if err != nil {
		return err
	}

	if !input.Reveal {
		for i := range configKeys {
			configKeys[i].Value = "***"
		}
	}

	w := os.Stdout

	if len(configKeys) == 0 {
		fmt.Fprintf(w, "No keys found for %s", cm.ServiceName())
		return nil
	}

	if input.Format == formatJSON {
		return outputJSON(w, cm.ServiceName(), configKeys)
	}

	tw := tabwriter.NewWriter(w, 25, 4, 2, ' ', 0)

	headers := append([]string{"Key", "Value"}, cm.MetadataKeys()...)
	headerUnderlining := make([]string, len(headers))
	for i, key := range headers {
		headerUnderlining[i] = strings.Repeat("=", len(key)+1)
	}

	fmt.Fprintf(tw, "Config for '%s'\n", cm.ServiceName())
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	fmt.Fprintln(tw, strings.Join(headerUnderlining, "\t"))

	for _, key := range configKeys {
		values := []string{key.Key, key.Value}
		for _, metadataKey := range cm.MetadataKeys() {
			values = append(values, key.Metadata[metadataKey])
		}

		fmt.Fprintf(tw, fmt.Sprintf("%s\n", strings.Join(values, "\t")))
	}

	return tw.Flush()
}
