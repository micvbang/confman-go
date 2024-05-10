package parameterstore

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/micvbang/go-helpy/inty"
	"github.com/micvbang/go-helpy/mapy"
	"github.com/prometheus/common/log"
)

// ParameterStore implements storage.Storage using Parameter Store from
// AWS Systems Manager
type ParameterStore struct {
	log       logger.Logger
	ssmClient SSMClient
	kmsKeyID  string
}

var _ storage.Storage = &ParameterStore{}

const kmsKeyAliasPrefix = "alias/"

type SSMClient interface {
	PutParameter(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)
	GetParameters(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error)
	GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
	DescribeParameters(ctx context.Context, params *ssm.DescribeParametersInput, optFns ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error)
	DeleteParameters(ctx context.Context, params *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error)
}

// New returns a configured instance of ParameterStore.
func New(log logger.Logger, ssmClient SSMClient, kmsKeyAlias string) *ParameterStore {
	if !strings.HasPrefix(kmsKeyAlias, kmsKeyAliasPrefix) {
		kmsKeyAlias = fmt.Sprintf("%s%s", kmsKeyAliasPrefix, kmsKeyAlias)
	}

	log = log.
		WithField("storage_type", "ParameterStore").
		WithField("kms_key", kmsKeyAlias)

	return &ParameterStore{
		log:       log,
		ssmClient: ssmClient,
		kmsKeyID:  kmsKeyAlias,
	}
}

func (ps *ParameterStore) Write(ctx context.Context, servicePath string, key string, value string) error {
	curValue, err := ps.Read(ctx, servicePath, key)
	if err != nil && !errors.Is(err, storage.ErrConfigNotFound) {
		return err
	}

	if curValue == value {
		return nil
	}

	_, err = ps.ssmClient.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(ps.parameterPath(servicePath, key)),
		KeyId:     aws.String(ps.kmsKeyID),
		Type:      types.ParameterTypeSecureString, // Encrypt all configuration
		Overwrite: aws.Bool(true),
		Value:     aws.String(value),

		// If compatible with segmentio/chamber, must write version number here.
		Description: aws.String(""),
	})
	return err
}

func (ps *ParameterStore) WriteKeys(ctx context.Context, servicePath string, config map[string]string) error {
	if len(config) == 0 {
		log.Warnf("WriteKeys called with 0 keys")
		return nil
	}

	keys := mapy.Keys(config)
	log.Debugf("Attempting to write keys %v %v", servicePath, keys)

	curConfig, err := ps.ReadKeys(ctx, servicePath, keys)
	if err != nil && !errors.Is(err, storage.ErrConfigNotFound) {
		return err
	}

	newValues := make(map[string]string, len(config))
	for key, newValue := range config {
		if newValue != curConfig[key] {
			newValues[key] = newValue
		}
	}

	for key, value := range newValues {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := ps.Write(ctx, servicePath, key, value)
		if err != nil {
			return err
		}
	}

	log.Debugf("Wrote keys %v %v", servicePath, keys)

	return nil
}

func (ps *ParameterStore) Read(ctx context.Context, servicePath string, key string) (value string, _ error) {
	config, err := ps.ReadKeys(ctx, servicePath, []string{key})
	if err != nil {
		return "", err
	}

	return config[key], nil
}

// maxKeysPerRequest is a limit set by AWS Parameter Store.
const maxKeysPerRequest = 10

func (ps *ParameterStore) ReadKeys(ctx context.Context, servicePath string, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		log.Warnf("ReadKeys called with 0 keys")
		return nil, nil
	}

	config := make(map[string]string, len(keys))

	for _, batchKeys := range ps.batchKeys(keys, maxKeysPerRequest) {
		batchConfig, err := ps.readKeys(ctx, servicePath, batchKeys)
		if err != nil {
			return nil, err
		}

		for key, value := range batchConfig {
			config[key] = value
		}
	}

	return config, nil
}

func (ps *ParameterStore) keyMetadataToKeyMetadata(readKey keyMetadata) storage.KeyMetadata {
	return storage.KeyMetadata{
		Key:   readKey.key,
		Value: readKey.value,
		Metadata: map[string]string{
			"description":        readKey.description,
			"version":            strconv.FormatInt(readKey.version, 10),
			"parameter_type":     readKey.parameterType,
			"last_modified_date": readKey.lastModifiedDate.Format(time.RFC3339),
			"last_modified_user": readKey.lastModifiedUser,
			"tier":               readKey.tier,
		},
	}
}

func (ps *ParameterStore) readKeys(ctx context.Context, servicePath string, keys []string) (map[string]string, error) {
	log := ps.populateLogger(servicePath)

	if len(keys) > maxKeysPerRequest {
		return nil, storage.ErrTooManyKeys
	}

	log.Debugf("Attempting to read keys %v", keys)

	output, err := ps.ssmClient.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          ps.keysToParameterNames(servicePath, keys),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "ParameterNotFound" {
				err = errors.Join(err, storage.ErrConfigNotFound)
			}
		}

		return nil, err
	}

	// Return no keys if just one did not exist.
	if len(output.InvalidParameters) > 0 {
		return nil, storage.ErrConfigNotFound
	}

	config := make(map[string]string, len(keys))
	for _, parameter := range output.Parameters {
		config[ps.parameterBaseName(parameter)] = aws.ToString(parameter.Value)
	}

	log.Debugf("Read keys %s %v", servicePath, keys)

	return config, nil
}

func (ps *ParameterStore) ReadAll(ctx context.Context, servicePath string) (map[string]string, error) {
	log := ps.populateLogger(servicePath)

	log.Debugf("Attempting to read all")

	config := make(map[string]string, 50)

	paginator := ssm.NewGetParametersByPathPaginator(ps.ssmClient, &ssm.GetParametersByPathInput{
		Path:             aws.String(servicePath),
		Recursive:        aws.Bool(false),
		WithDecryption:   aws.Bool(true),
		MaxResults:       aws.Int32(maxKeysPerRequest),
		ParameterFilters: nil,
		NextToken:        nil,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				if apiErr.ErrorCode() == "ParameterNotFound" {
					err = errors.Join(err, storage.ErrConfigNotFound)
				}
			}
			return nil, err
		}

		for _, p := range page.Parameters {
			key := ps.parameterBaseName(p)
			config[key] = aws.ToString(p.Value)
		}
	}

	return config, nil
}

func (ps *ParameterStore) ReadAllMetadata(ctx context.Context, servicePath string) ([]storage.KeyMetadata, error) {
	readKeys, err := ps.readAllKeyMetadata(ctx, servicePath)
	if err != nil {
		return nil, err
	}

	envConfigs := make([]storage.KeyMetadata, 0, len(readKeys))
	for _, readKey := range readKeys {
		envConfigs = append(envConfigs, ps.keyMetadataToKeyMetadata(readKey))
	}

	return envConfigs, nil
}

type keyMetadata struct {
	key              string
	value            string
	description      string
	version          int64
	parameterType    string
	lastModifiedDate time.Time
	lastModifiedUser string
	tier             string
}

func (ps *ParameterStore) readAllKeyMetadata(ctx context.Context, servicePath string) ([]keyMetadata, error) {
	log := ps.populateLogger(servicePath)

	log.Debugf("Attempting to read all metadata")

	config, err := ps.ReadAll(ctx, servicePath)
	if err != nil {
		return nil, err
	}

	keysMetadata := make([]keyMetadata, 0, 50)

	paginator := ssm.NewDescribeParametersPaginator(ps.ssmClient, &ssm.DescribeParametersInput{
		ParameterFilters: []types.ParameterStringFilter{
			{
				Key:    aws.String("Path"),
				Option: aws.String("OneLevel"),
				Values: []string{servicePath},
			},
		},
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				if apiErr.ErrorCode() == "ParameterNotFound" {
					err = errors.Join(err, storage.ErrConfigNotFound)
				}
			}

			return nil, err
		}

		for _, p := range page.Parameters {
			key := ps.parameterMetadataBaseName(p)

			keysMetadata = append(keysMetadata, keyMetadata{
				key:              key,
				value:            config[key],
				description:      aws.ToString(p.Description),
				version:          p.Version,
				lastModifiedDate: aws.ToTime(p.LastModifiedDate),
				parameterType:    string(p.Type),
				tier:             string(p.Tier),
				lastModifiedUser: aws.ToString(p.LastModifiedUser),
			})
		}
	}

	return keysMetadata, nil
}

func (ps *ParameterStore) Delete(ctx context.Context, servicePath string, key string) error {
	log := ps.populateLogger(servicePath)
	return ps.deleteKeys(ctx, log, servicePath, []string{key})
}

func (ps *ParameterStore) DeleteKeys(ctx context.Context, servicePath string, keys []string) error {
	log := ps.populateLogger(servicePath)

	if len(keys) == 0 {
		log.Warnf("DeleteKeys called with 0 keys")
		return nil
	}

	for _, batchKeys := range ps.batchKeys(keys, maxKeysPerRequest) {
		err := ps.deleteKeys(ctx, log, servicePath, batchKeys)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ps *ParameterStore) deleteKeys(ctx context.Context, log logger.Logger, servicePath string, keys []string) error {
	if len(keys) == 0 {
		log.Warnf("DeleteKeys called with 0 keys")
		return nil
	}

	log.Debugf("Attempting to delete keys %s %v", keys)

	if len(keys) > maxKeysPerRequest {
		return storage.ErrTooManyKeys
	}

	output, err := ps.ssmClient.DeleteParameters(ctx, &ssm.DeleteParametersInput{
		Names: ps.keysToParameterNames(servicePath, keys),
	})
	if err != nil {
		return err
	}

	if len(output.DeletedParameters) != len(keys) {
		log.Warnf("Attempted to delete %d keys, but deleted %v", len(keys), len(output.DeletedParameters))
	}

	for _, parameter := range output.DeletedParameters {
		log.Debugf("Deleted parameter %s", parameter)
	}

	// NOTE: not reporting errors when keys not deleted because.. well, they
	// aren't there now if they weren't found.
	for _, parameter := range output.InvalidParameters {
		log.Warnf("Failed to delete parameter %s", parameter)
	}

	return nil
}

func (ps *ParameterStore) parameterPath(servicePath string, key string) string {
	return path.Join(servicePath, key)
}

func (ps *ParameterStore) parameterMetadataBaseName(parameter types.ParameterMetadata) string {
	return path.Base(aws.ToString(parameter.Name))
}

func (ps *ParameterStore) parameterBaseName(parameter types.Parameter) string {
	return path.Base(aws.ToString(parameter.Name))
}

func (ps *ParameterStore) batchKeys(keys []string, batchSize int) [][]string {
	numBatches := (len(keys) / batchSize) + 1
	batches := make([][]string, 0, numBatches)

	for batchI := 0; batchI < numBatches; batchI++ {
		batch := make([]string, 0, batchSize)

		maxIters := inty.Min(len(keys), batchSize)
		for i := 0; i < maxIters; i++ {
			keyI := batchI*batchSize + i
			batch = append(batch, keys[keyI])
		}
		batches = append(batches, batch)
	}

	return batches
}

func (ps *ParameterStore) keysToParameterNames(servicePath string, keys []string) []string {
	names := make([]string, len(keys))

	for i, key := range keys {
		names[i] = ps.parameterPath(servicePath, key)
	}

	return names
}

func (ps *ParameterStore) String() string {
	return fmt.Sprintf("ParameterStore")
}

func (ps *ParameterStore) SetLogger(log logger.Logger) {
	ps.log = log
}

func (ps *ParameterStore) MetadataKeys() []string {
	return []string{
		"version",
		"last_modified_date",
		"last_modified_user",
	}
}

func (ps *ParameterStore) populateLogger(servicePath string) logger.Logger {
	return ps.log.WithField("service_path", servicePath)
}
