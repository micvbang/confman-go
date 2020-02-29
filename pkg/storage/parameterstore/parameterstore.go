package parameterstore

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"gitlab.com/micvbang/confman-go/pkg/storage"
)

// ParameterStore implements storage.Storage using Parameter Store from
// AWS Systems Manager
type ParameterStore struct {
	ssmClient   ssmiface.SSMAPI
	kmsKeyID    string
	serviceName string
}

var _ storage.Storage = &ParameterStore{}

// New returns a configured instance of ParameterStore.
func New(ssmClient ssmiface.SSMAPI, kmsKeyID string, serviceName string) *ParameterStore {
	// TODO: validate service name
	if !strings.HasPrefix(serviceName, "/") {
		serviceName = fmt.Sprintf("/serviceName")
	}

	return &ParameterStore{
		ssmClient:   ssmClient,
		kmsKeyID:    kmsKeyID,
		serviceName: serviceName,
	}
}

func (ps *ParameterStore) Add(ctx context.Context, key string, value string) error {
	curValue, err := ps.Read(ctx, key)
	if err != nil && err != storage.ErrConfigNotFound {
		return err
	}

	if curValue == value {
		return nil
	}

	_, err = ps.ssmClient.PutParameterWithContext(ctx, &ssm.PutParameterInput{
		Name:      aws.String(ps.parameterPath(key)),
		KeyId:     aws.String(ps.kmsKeyID),
		Type:      aws.String("SecureString"), // Encrypt all configuration
		Overwrite: aws.Bool(true),

		// If compatible with segmentio/chamber, must write version number here.
		Description: aws.String(""),
	})
	return err
}

func (ps *ParameterStore) Read(ctx context.Context, key string) (value string, _ error) {
	output, err := ps.ssmClient.GetParameterWithContext(ctx, &ssm.GetParameterInput{
		Name:           aws.String(ps.parameterPath(key)),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		if _, ok := err.(*ssm.ParameterNotFound); ok {
			return "", storage.ErrConfigNotFound
		}
		return "", err
	}

	return aws.StringValue(output.Parameter.Value), nil
}

func (ps *ParameterStore) ReadAll(ctx context.Context) (map[string]string, error) {
	const maxResults = 10
	config := make(map[string]string, 50)

	err := ps.ssmClient.GetParametersByPathPagesWithContext(ctx, &ssm.GetParametersByPathInput{
		Path:             aws.String(ps.serviceName),
		Recursive:        aws.Bool(true),
		WithDecryption:   aws.Bool(true),
		MaxResults:       aws.Int64(maxResults),
		ParameterFilters: nil,
		NextToken:        nil,
	}, func(output *ssm.GetParametersByPathOutput, b bool) bool {
		for _, p := range output.Parameters {
			key := path.Base(aws.StringValue(p.Name))
			config[key] = aws.StringValue(p.Value)
		}
		return true
	})

	if err != nil {
		if _, ok := err.(*ssm.ParameterNotFound); ok {
			return nil, storage.ErrConfigNotFound
		}
		return nil, err
	}

	return config, nil
}

func (ps *ParameterStore) Delete(ctx context.Context, key string) error {
	return nil
}

func (ps *ParameterStore) DeleteKeys(ctx context.Context, keys []string) error {
	return nil
}

func (ps *ParameterStore) parameterPath(key string) string {
	return path.Join(ps.serviceName, key)
}
