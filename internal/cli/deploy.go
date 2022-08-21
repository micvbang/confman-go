package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/micvbang/confman-go/pkg/configuration"
	"github.com/micvbang/confman-go/pkg/confman"
	"github.com/micvbang/go-helpy/mapy"

	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type DeployCommandInput struct {
	Path      string
	Base      string
	Define    bool
	AssumeYes bool
}

func ConfigureDeployCommand(ctx context.Context, app *kingpin.Application, log logger.Logger) {
	input := DeployCommandInput{}

	cmd := app.Command("deploy", "deploys configuration from a file")
	cmd.Arg("path", "File or folder containing configuration").
		Required().
		ExistingFileOrDirVar(&input.Path)

	baseDefault, _ := os.Getwd()
	cmd.Flag("base", "Base path from which to determine service paths").
		Default(baseDefault).
		StringVar(&input.Base)

	cmd.Flag("define", "Whether to 'define' the configuration, i.e. removing keys from the store which don't exist in the given configuration").
		Default("false").
		BoolVar(&input.Define)

	cmd.Flag("assume-yes", "Assume yes to key deletion prompts (used with \"--define\")").
		Short('y').
		Default("false").
		BoolVar(&input.AssumeYes)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(DeployCommand(ctx, input, os.Stdin, os.Stdout, log, GlobalFlags.Storage), "deploy")
		return nil
	})
}

func DeployCommand(ctx context.Context, input DeployCommandInput, rd io.Reader, w io.Writer, log logger.Logger, storage storage.Storage) error {
	serviceConfigs, err := configuration.Read(input.Path)
	if err != nil {
		return err
	}

	for _, serviceConfig := range serviceConfigs {
		servicePath := filePathToServicePath(input.Base, serviceConfig.Path)
		cm := confman.New(log, storage, servicePath)

		if input.Define {
			err = handleDefine(ctx, cm, rd, w, serviceConfig.Config, input.AssumeYes)
			if err != nil {
				return err
			}
		}

		for key, value := range serviceConfig.Config {
			fmt.Fprintf(w, "Updating %v %v = %v\n", servicePath, key, value)
		}
		err := cm.WriteKeys(ctx, serviceConfig.Config)
		if err != nil {
			return err
		}
	}
	return nil
}

var ErrUserAbortedKeyDeletion = errors.New("user aborted key deletion")

func handleDefine(ctx context.Context, cm confman.Confman, rd io.Reader, w io.Writer, newConfig map[string]string, assumeYes bool) error {
	existingConfig, err := cm.ReadAll(ctx)
	if err != nil {
		return err
	}

	deleteKeys := []string{}
	existingConfigKeys := mapy.Keys(existingConfig)
	for _, configKey := range existingConfigKeys {
		if _, contains := newConfig[configKey]; !contains {
			deleteKeys = append(deleteKeys, configKey)
		}
	}

	if len(deleteKeys) == 0 {
		return nil
	}

	if !assumeYes {
		fmt.Fprintf(w, "Are you sure that you want to delete the following keys: %v?\nyes/no: ", deleteKeys)
		var userInput string
		_, err = fmt.Fscanln(rd, &userInput)
		if err != nil {
			return err
		}

		userInput = strings.ToLower(userInput)
		if !strings.HasPrefix(userInput, "y") {
			return ErrUserAbortedKeyDeletion
		}
	}

	return cm.DeleteKeys(ctx, deleteKeys)
}

func filePathToServicePath(base string, filePath string) string {
	servicePath := strings.TrimPrefix(filePath, base)
	ext := filepath.Ext(filePath)
	return servicePath[:len(servicePath)-len(ext)]
}
