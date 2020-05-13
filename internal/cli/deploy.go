package cli

import (
	"context"
	"fmt"
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
	Path            string
	Base            string
	Define          bool
	DefineYesDelete bool
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

	cmd.Flag("yes", "Whether to answer \"yes\" to prompts for key deletion (used with \"--define\")").
		Short('y').
		Default("false").
		BoolVar(&input.DefineYesDelete)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(DeployCommand(ctx, app, input, log, GlobalFlags.Storage), "deploy")
		return nil
	})
}

func DeployCommand(ctx context.Context, app *kingpin.Application, input DeployCommandInput, log logger.Logger, storage storage.Storage) error {
	serviceConfigs, err := configuration.Read(input.Path)
	if err != nil {
		return err
	}

	for _, serviceConfig := range serviceConfigs {
		servicePath := filePathToServicePath(input.Base, serviceConfig.Path)
		cm := confman.New(log, storage, servicePath)

		if input.Define {
			err = handleDefine(ctx, cm, serviceConfig.Config, input.DefineYesDelete)
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

func handleDefine(ctx context.Context, cm confman.Confman, newConfig map[string]string, yesDelete bool) error {
	existingConfig, err := cm.ReadAll(ctx)
	if err != nil {
		return err
	}

	existingConfigKeys, _ := mapy.StringKeys(existingConfig)
	deleteKeys := []string{}

	for _, configKey := range existingConfigKeys {
		if _, contains := newConfig[configKey]; !contains {
			deleteKeys = append(deleteKeys, configKey)
		}
	}

	if len(deleteKeys) == 0 {
		return nil
	}

	if !yesDelete {
		fmt.Printf("Are you sure that you want to delete the following keys: %v?\nyes/no: ", deleteKeys)
		var userInput string
		_, err = fmt.Fscanln(os.Stdin, &userInput)
		if err != nil {
			return err
		}

		userInput = strings.ToLower(userInput)
		if !strings.HasPrefix(userInput, "y") {
			return fmt.Errorf("user aborted key deletion in define")
		}
	}

	return cm.DeleteKeys(ctx, deleteKeys)
}

func filePathToServicePath(base string, filePath string) string {
	servicePath := strings.TrimPrefix(filePath, base)
	ext := filepath.Ext(filePath)
	return servicePath[:len(servicePath)-len(ext)]
}
