package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"text/tabwriter"

	"github.com/micvbang/confman-go/pkg/confman"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ListCommandInput struct {
	ServiceNames string
	Key          string
	Format       string
	Reveal       bool
	Quiet        bool
}

func ConfigureListCommand(ctx context.Context, app *kingpin.Application, log logger.Logger) {
	input := ListCommandInput{}

	cmd := app.Command("list", "Lists configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServiceNames)

	cmd.Flag("reveal", "Reveal values").
		Envar("CONFMAN_REVEAL_VALUES").
		BoolVar(&input.Reveal)

	addFlagOutputFormat(cmd, &input.Format)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ListCommand(ctx, app, input, log, GlobalFlags.Storage), "list")
		return nil
	})
}

func ListCommand(ctx context.Context, app *kingpin.Application, input ListCommandInput, log logger.Logger, s storage.Storage) error {
	serviceConfigKeys := make(map[string][]storage.KeyMetadata)

	var metadataKeys []string
	serviceNames := confman.ParseServicePaths(input.ServiceNames)
	for _, serviceName := range serviceNames {
		cm := confman.New(log, s, serviceName)
		configKeys, err := cm.ReadAllMetadata(ctx)
		if err != nil {
			return err
		}
		serviceConfigKeys[serviceName] = configKeys
		metadataKeys = cm.MetadataKeys()
	}

	if !input.Reveal {
		for serviceName, configKeys := range serviceConfigKeys {
			for i := range configKeys {
				serviceConfigKeys[serviceName][i].Value = "***"
			}
		}
	}

	w := os.Stdout

	if input.Format != formatText {
		return outputFormat(input.Format, w, serviceConfigKeys)
	}

	tw := tabwriter.NewWriter(w, 25, 4, 2, ' ', 0)

	i := 0
	numNewlines := len(serviceConfigKeys) - 1
	for serviceName, configKeys := range serviceConfigKeys {
		headers := append([]string{"Key", "Value"}, metadataKeys...)
		headerUnderlining := make([]string, len(headers))
		for i, key := range headers {
			headerUnderlining[i] = strings.Repeat("=", len(key)+1)
		}

		fmt.Fprintf(tw, "Config for '%s'\n", serviceName)
		fmt.Fprintln(tw, strings.Join(headers, "\t"))
		fmt.Fprintln(tw, strings.Join(headerUnderlining, "\t"))

		for _, key := range configKeys {
			values := []string{key.Key, key.Value}
			for _, metadataKey := range metadataKeys {
				values = append(values, key.Metadata[metadataKey])
			}

			fmt.Fprintf(tw, fmt.Sprintf("%s\n", strings.Join(values, "\t")))
		}

		if i < numNewlines {
			fmt.Fprint(tw, "\n")
		}
		i++
	}

	return tw.Flush()
}
