package parameterstore

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/micvbang/go-helpy/inty"
	"github.com/micvbang/go-helpy/mapy"
)

// ParameterStore implements storage.Storage using Parameter Store from
// AWS Systems Manager
type ParameterStore struct {
	log       logger.Logger
	ssmClient ssmiface.SSMAPI
	kmsKeyID  string
}

var _ storage.Storage = &ParameterStore{}

const kmsKeyAliasPrefix = "alias/"

// New returns a configured instance of ParameterStore.
func New(log logger.Logger, ssmClient ssmiface.SSMAPI, kmsKeyAlias string) *ParameterStore {
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

func (ps *ParameterStore) Write(ctx context.Context, serviceName string, key string, value string) error {
	curValue, err := ps.Read(ctx, serviceName, key)
	if err != nil && err != storage.ErrConfigNotFound {
		return err
	}

	if curValue == value {
		return nil
	}

	_, err = ps.ssmClient.PutParameterWithContext(ctx, &ssm.PutParameterInput{
		Name:      aws.String(ps.parameterPath(serviceName, key)),
		KeyId:     aws.String(ps.kmsKeyID),
		Type:      aws.String("SecureString"), // Encrypt all configuration
		Overwrite: aws.Bool(true),
		Value:     aws.String(value),

		// If compatible with segmentio/chamber, must write version number here.
		Description: aws.String(""),
	})
	return err
}

func (ps *ParameterStore) WriteKeys(ctx context.Context, serviceName string, config map[string]string) error {
	if len(config) == 0 {
		ps.log.Warnf("WriteKeys called with 0 keys")
		return nil
	}

	keys, _ := mapy.StringKeys(config)
	ps.log.Debugf("Attempting to write keys %v", keys)

	curConfig, err := ps.ReadKeys(ctx, serviceName, keys)
	if err != nil && err != storage.ErrConfigNotFound {
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

		err := ps.Write(ctx, serviceName, key, value)
		if err != nil {
			return err
		}
	}

	ps.log.Debugf("Wrote keys %v", keys)

	return nil
}

func (ps *ParameterStore) Read(ctx context.Context, serviceName string, key string) (value string, _ error) {
	config, err := ps.ReadKeys(ctx, serviceName, []string{key})
	if err != nil {
		return "", err
	}

	return config[key], nil
}

// maxKeysPerRequest is a limit set by AWS Parameter Store.
const maxKeysPerRequest = 10

func (ps *ParameterStore) ReadKeys(ctx context.Context, serviceName string, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		ps.log.Warnf("ReadKeys called with 0 keys")
		return nil, nil
	}

	config := make(map[string]string, len(keys))

	for _, batchKeys := range ps.batchKeys(keys, maxKeysPerRequest) {
		batchConfig, err := ps.readKeys(ctx, serviceName, batchKeys)
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

func (ps *ParameterStore) readKeys(ctx context.Context, serviceName string, keys []string) (map[string]string, error) {
	if len(keys) > maxKeysPerRequest {
		return nil, storage.ErrTooManyKeys
	}

	ps.log.Debugf("Attempting to read keys %v", keys)

	output, err := ps.ssmClient.GetParametersWithContext(ctx, &ssm.GetParametersInput{
		Names:          ps.keysToParameterNames(serviceName, keys),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		if _, ok := err.(*ssm.ParameterNotFound); ok {
			return nil, storage.ErrConfigNotFound
		}
		return nil, err
	}

	// Return no keys if just one did not exist.
	if len(output.InvalidParameters) > 0 {
		return nil, storage.ErrConfigNotFound
	}

	config := make(map[string]string, len(keys))
	for _, parameter := range output.Parameters {
		config[ps.parameterBaseName(parameter)] = aws.StringValue(parameter.Value)
	}

	ps.log.Debugf("Read keys %v", keys)

	return config, nil
}

func (ps *ParameterStore) ReadAll(ctx context.Context, serviceName string) (map[string]string, error) {
	log := ps.log.WithField("service_name", serviceName)

	log.Debugf("Attempting to read all")

	config := make(map[string]string, 50)

	err := ps.ssmClient.GetParametersByPathPagesWithContext(ctx, &ssm.GetParametersByPathInput{
		Path:             aws.String(serviceName),
		Recursive:        aws.Bool(false),
		WithDecryption:   aws.Bool(true),
		MaxResults:       aws.Int64(maxKeysPerRequest),
		ParameterFilters: nil,
		NextToken:        nil,
	}, func(output *ssm.GetParametersByPathOutput, b bool) bool {
		for _, p := range output.Parameters {
			key := ps.parameterBaseName(p)
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

func (ps *ParameterStore) ReadAllMetadata(ctx context.Context, serviceName string) ([]storage.KeyMetadata, error) {
	readKeys, err := ps.readAllKeyMetadata(ctx, serviceName)
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

func (ps *ParameterStore) readAllKeyMetadata(ctx context.Context, serviceName string) ([]keyMetadata, error) {
	log := ps.log.WithField("service_name", serviceName)

	log.Debugf("Attempting to read all metadata")

	config, err := ps.ReadAll(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	keysMetadata := make([]keyMetadata, 0, 50)

	err = ps.ssmClient.DescribeParametersPagesWithContext(ctx, &ssm.DescribeParametersInput{
		ParameterFilters: []*ssm.ParameterStringFilter{
			&ssm.ParameterStringFilter{
				Key:    aws.String("Path"),
				Option: aws.String("OneLevel"),
				Values: []*string{aws.String(serviceName)},
			},
		},
	}, func(output *ssm.DescribeParametersOutput, b bool) bool {
		for _, p := range output.Parameters {
			key := ps.parameterMetadataBaseName(p)

			keysMetadata = append(keysMetadata, keyMetadata{
				key:              key,
				value:            config[key],
				description:      aws.StringValue(p.Description),
				version:          aws.Int64Value(p.Version),
				lastModifiedDate: aws.TimeValue(p.LastModifiedDate),
				parameterType:    aws.StringValue(p.Type),
				tier:             aws.StringValue(p.Tier),
				lastModifiedUser: aws.StringValue(p.LastModifiedUser),
			})
		}

		return true
	})
	if err != nil {
		if _, ok := err.(*ssm.ParameterNotFound); ok {
			return nil, storage.ErrConfigNotFound
		}
		return nil, err
	}

	return keysMetadata, nil
}

func (ps *ParameterStore) Delete(ctx context.Context, serviceName string, key string) error {
	return ps.deleteKeys(ctx, serviceName, []string{key})
}

func (ps *ParameterStore) DeleteKeys(ctx context.Context, serviceName string, keys []string) error {
	if len(keys) == 0 {
		ps.log.Warnf("DeleteKeys called with 0 keys")
		return nil
	}

	for _, batchKeys := range ps.batchKeys(keys, maxKeysPerRequest) {
		err := ps.deleteKeys(ctx, serviceName, batchKeys)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ps *ParameterStore) deleteKeys(ctx context.Context, serviceName string, keys []string) error {
	log := ps.log.WithField("service_name", serviceName)

	if len(keys) == 0 {
		ps.log.Warnf("DeleteKeys called with 0 keys")
		return nil
	}

	log.Debugf("Attempting to delete keys %s %v", keys)

	if len(keys) > maxKeysPerRequest {
		return storage.ErrTooManyKeys
	}

	output, err := ps.ssmClient.DeleteParametersWithContext(ctx, &ssm.DeleteParametersInput{
		Names: ps.keysToParameterNames(serviceName, keys),
	})
	if err != nil {
		return err
	}

	if len(output.DeletedParameters) != len(keys) {
		log.Warnf("Attempted to delete %d keys, but deleted %v", len(keys), len(output.DeletedParameters))
	}

	for _, parameter := range output.DeletedParameters {
		log.Debugf("Deleted parameter %s", aws.StringValue(parameter))
	}

	// NOTE: not reporting errors when keys not deleted because.. well, they
	// aren't there now if they weren't found.
	for _, parameter := range output.InvalidParameters {
		log.Warnf("Failed to delete parameter %s", aws.StringValue(parameter))
	}

	return nil
}

func (ps *ParameterStore) parameterPath(serviceName string, key string) string {
	return path.Join(serviceName, key)
}

func (ps *ParameterStore) parameterMetadataBaseName(parameter *ssm.ParameterMetadata) string {
	// TODO: don't fail silently like this.
	if parameter == nil {
		return ""
	}

	return path.Base(aws.StringValue(parameter.Name))
}

func (ps *ParameterStore) parameterBaseName(parameter *ssm.Parameter) string {
	// TODO: don't fail silently like this.
	if parameter == nil {
		return ""
	}

	return path.Base(aws.StringValue(parameter.Name))
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

func (ps *ParameterStore) keysToParameterNames(serviceName string, keys []string) []*string {
	names := make([]*string, len(keys))

	for i, key := range keys {
		names[i] = aws.String(ps.parameterPath(serviceName, key))
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
