package parameterstore

import (
	"context"
	"path"

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
	return &ParameterStore{
		ssmClient:   ssmClient,
		kmsKeyID:    kmsKeyID,
		serviceName: serviceName,
	}
}

func (ps *ParameterStore) Add(ctx context.Context, key string, value string) error {
	curValue, err := ps.Read(ctx, key)
	if err != nil {
		return err
	}

	if curValue == value {
		return nil
	}

	ps.ssmClient.PutParameterWithContext(ctx, &ssm.PutParameterInput{
		Name:      aws.String(ps.parameterPath(key)),
		KeyId:     aws.String(ps.kmsKeyID),
		Type:      aws.String("SecureString"), // Encrypt all configuration
		Overwrite: aws.Bool(true),

		// If compatible with segmentio/chamber, must write version number here.
		Description: aws.String(""),
	})
	return nil
}

func (ps *ParameterStore) Read(ctx context.Context, key string) (value string, _ error) {
	output, err := ps.ssmClient.GetParameterWithContext(ctx, &ssm.GetParameterInput{
		Name:           aws.String(ps.parameterPath(key)),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}

	return aws.StringValue(output.Parameter.Value), nil
}

func (ps *ParameterStore) ReadAll(ctx context.Context) (map[string]string, error) {
	return ps.readAll(ctx, nil)
}

func (ps *ParameterStore) readAll(ctx context.Context, nextToken *string) (map[string]string, error) {
	const maxResults = 10
	parameters := make(map[string]string, 50)

	numResults := maxResults
	for numResults >= maxResults {
		output, err := ps.ssmClient.GetParametersByPathWithContext(ctx, &ssm.GetParametersByPathInput{
			Path:             aws.String(ps.serviceName),
			Recursive:        aws.Bool(true),
			WithDecryption:   aws.Bool(true),
			MaxResults:       aws.Int64(maxResults),
			ParameterFilters: nil,
			NextToken:        nextToken,
		})
		if err != nil {
			return nil, err
		}
		numResults = len(output.Parameters)

		for _, p := range output.Parameters {
			key := path.Base(aws.StringValue(p.Name))
			parameters[key] = aws.StringValue(p.Value)
		}
	}

	return parameters, nil
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
