package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	builtinLog "log"

	"github.com/99designs/aws-vault/v6/vault"
	"github.com/99designs/keyring"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/micvbang/confman-go/pkg/confman"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/micvbang/confman-go/pkg/storage/parameterstore"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// TODO: make not global
var storageBackend storage.Storage

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
				builtinLog.SetOutput(ioutil.Discard)
			}

			logrusLog.Level = logrus.WarnLevel
		}

		session, err := makeAWSSession(GlobalFlags.AssumeProfile)
		if err != nil {
			app.Fatalf("Failed to start AWS session: %v", err)
		}

		confman.ChamberCompatible = GlobalFlags.ChamberCompatible
		ssmClient := ssm.New(session)
		GlobalFlags.Storage = parameterstore.New(log, ssmClient, "parameter_store_key")

		return nil
	})

	return log
}

func makeAWSSession(assumeProfile string) (*session.Session, error) {
	if len(assumeProfile) == 0 {
		return session.NewSession(&aws.Config{
			Credentials: credentials.NewEnvCredentials(),
		})
	}

	// Attempt to assume the given profile
	configFile, err := vault.LoadConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %v", err)
	}

	configLoader := vault.ConfigLoader{
		File:          configFile,
		BaseConfig:    vault.Config{},
		ActiveProfile: assumeProfile,
	}
	config, err := configLoader.LoadFromProfile(assumeProfile)
	if err != nil {
		return nil, err
	}

	keyring, err := keyring.Open(keyring.Config{
		ServiceName: "aws-vault",
		FileDir:     "~/.awsvault/keys/",
		// FilePasswordFunc:         fileKeyringPassphrasePrompt,
		LibSecretCollectionName:  "awsvault",
		KWalletAppID:             "aws-vault",
		KWalletFolder:            "aws-vault",
		KeychainTrustApplication: true,
		WinCredPrefix:            "aws-vault",
	})
	if err != nil {
		return nil, err
	}

	ckr := &vault.CredentialKeyring{Keyring: keyring}
	creds, err := vault.NewTempCredentialsProvider(config, ckr)
	if err != nil {
		return nil, fmt.Errorf("error getting temporary credentials: %w", err)
	}

	val, err := creds.Retrieve(context.Background())
	if err != nil {
		return nil, err
	}

	return session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(val.AccessKeyID, val.SecretAccessKey, val.SessionToken),
	})
}
