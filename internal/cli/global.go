package cli

import (
	"context"
	"fmt"
	"io"
	builtinLog "log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/micvbang/confman-go/pkg/confman"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/micvbang/confman-go/pkg/storage/parameterstore"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/ini.v1"
)

var GlobalFlags struct {
	Debug bool

	// TODO: only do for storage backends that need it
	AWSRegion         string
	KMSKeyAlias       string
	ChamberCompatible bool
	AssumeProfile     string

	Storage storage.Storage
}

func ConfigureGlobals(app *kingpin.Application) logger.Logger {
	logrusLog := logrus.New()
	log := logger.LogrusWrapper{Logger: logrusLog}

	app.Flag("debug", "Show debugging output").
		BoolVar(&GlobalFlags.Debug)

	app.Flag("aws-kms-key-alias", "KMS key alias used for config en/decryption").
		Default("parameter_store_key").
		Envar("CONFMAN_KMS_KEY_ALIAS").
		StringVar(&GlobalFlags.KMSKeyAlias)

	app.Flag("chamber-compatible", "Read and write data in a way that is compatible with chamber").
		Default("false").
		Envar("CONFMAN_CHAMBER_COMPATIBLE").
		BoolVar(&GlobalFlags.ChamberCompatible)

	app.Flag("assume-profile", "Attempt to assume the given AWS profile").
		Default("").
		Envar("CONFMAN_ASSUME_PROFILE").
		StringVar(&GlobalFlags.AssumeProfile)

	// TODO: determine storage backend from env/flags
	// TODO: determine AWS config from env/flags

	app.PreAction(func(c *kingpin.ParseContext) (err error) {
		if GlobalFlags.Debug {
			logrusLog.Level = logrus.DebugLevel
		} else {
			if len(GlobalFlags.AssumeProfile) > 0 {
				// For silencing logs from aws-vault
				builtinLog.SetOutput(io.Discard)
			}

			logrusLog.Level = logrus.WarnLevel
		}

		awsCfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to init aws config: %v", err)
		}

		if len(GlobalFlags.AssumeProfile) > 0 {
			roleARN, err := getAWSProfileRoleARN(GlobalFlags.AssumeProfile)
			if err != nil {
				return err
			}

			stsClient := sts.NewFromConfig(awsCfg)
			assumeRoleProvider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)
			awsCfg.Credentials = aws.NewCredentialsCache(assumeRoleProvider)
		}

		confman.ChamberCompatible = GlobalFlags.ChamberCompatible
		ssmClient := ssm.NewFromConfig(awsCfg)
		GlobalFlags.Storage = parameterstore.New(log, ssmClient, "parameter_store_key")

		return nil
	})

	return log
}

func getAWSProfileRoleARN(profileName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("looking up homedir: %w", err)
	}

	awsConfigPath := filepath.Join(homeDir, ".aws/config")
	f, err := ini.Load(awsConfigPath)
	if err != nil {
		return "", fmt.Errorf("reading aws config at '%s': %w", awsConfigPath, err)
	}

	profileNames := []string{
		fmt.Sprintf("profile %s", profileName),
		profileName,
	}

	const roleARNKey = "role_arn"

	for _, profileName := range profileNames {
		section := f.Section(profileName)
		if section == nil {
			continue
		}
		key := section.Key(roleARNKey)
		if key == nil {
			continue
		}

		return key.String(), nil
	}

	return "", fmt.Errorf("profile '%s' not found", profileName)
}
