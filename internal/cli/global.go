package cli

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/sirupsen/logrus"
	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/logger"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gitlab.com/micvbang/confman-go/pkg/storage/parameterstore"
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

	// TODO: determine storage backend from env/flags
	// TODO: determine AWS config from env/flags
	session, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		app.Fatalf("Failed to start AWS session: %v", err)
	}

	app.PreAction(func(c *kingpin.ParseContext) (err error) {
		if GlobalFlags.Debug {
			logrusLog.Level = logrus.DebugLevel
		} else {
			logrusLog.Level = logrus.WarnLevel
		}

		confman.ChamberCompatible = GlobalFlags.ChamberCompatible
		ssmClient := ssm.New(session)
		GlobalFlags.Storage = parameterstore.New(log, ssmClient, "parameter_store_key")

		return nil
	})

	return log
}
