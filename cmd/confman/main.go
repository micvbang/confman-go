package main

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/sirupsen/logrus"
	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/storage/parameterstore"
)

func main() {
	logrusLog := logrus.New()
	logrusLog.Level = logrus.DebugLevel
	log := confman.LogrusWrapper{logrusLog}

	const (
		awsRegion    = "eu-central-1"
		kmsKeyAlias  = "parameter_store_key"
		serviceName  = "/test-delete-me/"
		serviceName2 = "/test-delete-me-2/"
	)

	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		log.Fatalf("Failed to start AWS session: %v", err)
	}

	ssmClient := ssm.New(session)
	cm2 := confman.New(log, parameterstore.New(log, ssmClient, kmsKeyAlias, serviceName2))
	cm := confman.New(log, parameterstore.New(log, ssmClient, kmsKeyAlias, serviceName))

	ctx := context.Background()
	err = cm.Move(ctx, cm2)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	// cm2.Move(ctx, cm)
	// err = cm.AddKeys(ctx, map[string]string{
	// 	"test-key-1": "test-value-1",
	// 	"test-key-2": "test-value-2",
	// 	"test-key-3": "test-value-3",
	// })
	// if err != nil {
	// 	log.Fatalf(err.Error())
	// }

	// config, err := cm.ReadAll(ctx)
	// if err != nil {
	// 	log.Fatalf(err.Error())
	// }

	// for k, v := range config {
	// 	log.Printf("%v: %v", k, v)
	// }
}
