package cli

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/sirupsen/logrus"
	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gitlab.com/micvbang/confman-go/pkg/storage/parameterstore"
	"gopkg.in/alecthomas/kingpin.v2"
)

// TODO: make not global
var storageBackend storage.Storage

var GlobalFlags struct {
	Debug bool

	// TODO: only do for storage backends that need it
	AWSRegion   string
	KMSKeyAlias string
}

func ConfigureGlobals(app *kingpin.Application) (confman.Logger, storage.Storage) {
	logrusLog := logrus.New()
	log := confman.LogrusWrapper{Logger: logrusLog}

	app.Flag("debug", "Show debugging output").
		BoolVar(&GlobalFlags.Debug)

	// TODO: only configure for storage backends that need it
	app.Flag("aws-region", "AWS region").
		Envar("AWS_REGION").
		StringVar(&GlobalFlags.AWSRegion)

	app.Flag("aws-kms-key-alias", "KMS key alias used for config en/decryption").
		Default("parameter_store_key").
		Envar("CONFMAN_KMS_KEY_ALIAS").
		StringVar(&GlobalFlags.KMSKeyAlias)

	var storage storage.Storage
	// TODO: determine storage backend from env/flags
	// TODO: determine AWS config from env/flags
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(GlobalFlags.AWSRegion),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		app.Fatalf("Failed to start AWS session: %v", err)
	}

	ssmClient := ssm.New(session)
	storage = parameterstore.New(log, ssmClient, "parameter_store_key")

	app.PreAction(func(c *kingpin.ParseContext) (err error) {
		if GlobalFlags.Debug {
			logrusLog.Level = logrus.DebugLevel
		} else {
			logrusLog.Level = logrus.PanicLevel
		}

		return nil
	})

	return log, storage
}
