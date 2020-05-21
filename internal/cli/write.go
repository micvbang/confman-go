package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/99designs/aws-vault/prompt"
	"github.com/micvbang/confman-go/pkg/logger"

	"github.com/micvbang/confman-go/pkg/confman"
	"github.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type WriteCommandInput struct {
	ServicePath string
	Key         string
	Value       string
	Format      string
}

func ConfigureWriteCommand(ctx context.Context, app *kingpin.Application, log logger.Logger) {
	input := WriteCommandInput{}

	cmd := app.Command("write", "Writes a configuration")
	cmd.Arg("service", "Name of the service").
		Required().
		StringVar(&input.ServicePath)

	cmd.Arg("key", "Name of the key").
		Required().
		StringVar(&input.Key)

	addFlagOutputFormat(cmd, &input.Format)

	cmd.Flag("value", "Value to write (don't use this flag for secret values)").
		Short('v').
		StringVar(&input.Value)

	cmd.Action(func(c *kingpin.ParseContext) error {

		app.FatalIfError(WriteCommand(ctx, input, os.Stdout, log, GlobalFlags.Storage), "write")
		return nil
	})
}

func WriteCommand(ctx context.Context, input WriteCommandInput, w io.Writer, log logger.Logger, storage storage.Storage) error {
	cm := confman.New(log, storage, input.ServicePath)

	var err error
	value := input.Value
	if len(value) == 0 {
		value, err = prompt.TerminalPrompt(fmt.Sprintf("Enter value for key '%s': ", input.Key))
		if err != nil {
			return err
		}
	}

	err = cm.Write(ctx, input.Key, value)
	if err != nil {
		return err
	}

	if input.Format != formatText {
		return outputFormat(input.Format, w, map[string]interface{}{
			cm.ServicePath(): map[string]string{
				input.Key: value,
			},
		})
	}

	fmt.Fprintf(w, "%s = '%s'", cm.FormatKeyPath(input.Key), value)
	return nil
}
