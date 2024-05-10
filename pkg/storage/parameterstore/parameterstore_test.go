package parameterstore_test

import (
	"context"
	"fmt"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/micvbang/confman-go/pkg/storage/parameterstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	log                  = logger.LogrusWrapper{Logger: logrus.New()}
	ErrParameterNotFound = &smithy.GenericAPIError{Code: "ParameterNotFound"}
)

// TestParameterStoreReadExists verifies that ParameterStore returns the
// expected value when the key to be read exists in AWS Parameter Store.
func TestParameterStoreReadExists(t *testing.T) {
	const (
		servicePath   = "/service/env"
		key           = "key"
		expectedValue = "the value"
	)

	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.MockGetParameters = func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
		parameters := makeParameters(servicePath, map[string]string{
			key: expectedValue,
		})

		return &ssm.GetParametersOutput{Parameters: parameters}, nil
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	gotValue, err := ps.Read(ctx, servicePath, key)

	require.NoError(t, err)
	require.Equal(t, expectedValue, gotValue)
}

// TestParameterStoreReadNotExists verifies that ErrConfigNotFound is returned
// when the given key does not exist.
func TestParameterStoreReadNotExists(t *testing.T) {
	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.MockGetParameters = func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
		return nil, ErrParameterNotFound
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	_, err := ps.Read(ctx, "/some/service", "some-key")

	require.ErrorIs(t, err, storage.ErrConfigNotFound)
}

// TestParameterStoreWriteExistsNotEqual verifies that Put updates the given key
// in AWS Parameter Store when the key already exists, but the stored value is
// not equal to the input.
func TestParameterStoreWriteExistsNotEqual(t *testing.T) {
	const (
		servicePath = "/service/env"
		key         = "key"
		value       = "the value"
	)

	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.MockGetParameters = func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
		parameters := makeParameters(servicePath, map[string]string{key: "not value"})
		return &ssm.GetParametersOutput{Parameters: parameters}, nil
	}
	ssmMock.MockPutParameter = func(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
		return &ssm.PutParameterOutput{}, nil
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	err := ps.Write(ctx, servicePath, key, value)

	require.NoError(t, err)
	require.True(t, ssmMock.PutParameterCalled)
}

// TestParameterStoreWriteExistsEqual verifies that Put does not update the
// given key in AWS Parameter Store when the key/value pair already exists,
// and the stored value is equal to the input.
func TestParameterStoreWriteExistsEqual(t *testing.T) {
	const (
		servicePath = "/service/env"
		key         = "key"
		value       = "the value"
	)

	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.MockGetParameters = func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
		parameters := makeParameters(servicePath, map[string]string{key: value})
		return &ssm.GetParametersOutput{Parameters: parameters}, nil
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	err := ps.Write(ctx, servicePath, key, value)

	require.NoError(t, err)
	require.True(t, ssmMock.GetParametersCalled)
	require.False(t, ssmMock.PutParameterCalled)
}

// TestParameterStoreWriteNotExists verifies that Put Writes the given
// key/value pair to AWS Parameter Store when the given key does not already
// exist.
func TestParameterStoreWriteNotExists(t *testing.T) {
	const (
		servicePath = "/service/env"
		key         = "key"
		value       = "the value"
	)

	ssmMock := &parameterstore.MockSSMClient{}

	ssmMock.MockGetParameters = func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
		return nil, ErrParameterNotFound
	}

	ssmMock.MockPutParameter = func(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
		return &ssm.PutParameterOutput{}, nil
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	err := ps.Write(ctx, servicePath, key, value)

	require.NoError(t, err)
	require.True(t, ssmMock.GetParametersCalled)
	require.True(t, ssmMock.PutParameterCalled)
}

// TestParameterStoreReadAllServiceExists verifies that configs are retrieved
// correctly for services with a single page worth of AWS Parameter Store
// parameters.
func TestParameterStoreReadAllServiceExists(t *testing.T) {
	const servicePath = "/service-name"

	ssmMock := &parameterstore.MockSSMClient{}

	expectedConfig := map[string]string{
		"var1": "val1",
		"var2": "val2",
		"var3": "val3",
	}

	mockGetParametersByPathPagesWithContext(ssmMock, servicePath, expectedConfig)

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	gotConfig, err := ps.ReadAll(ctx, servicePath)
	require.NoError(t, err)

	requireConfigEqual(t, expectedConfig, gotConfig)
}

// TestParameterStoreReadAllServiceExistsMultiplePages verifies that configs
// are retrieved corrcetly for services with multiple AWS Parameter Store pages
// worth of parameters.
func TestParameterStoreReadAllServiceExistsMultiplePages(t *testing.T) {
	const servicePath = "/service-name"

	ssmMock := &parameterstore.MockSSMClient{}

	config1 := map[string]string{
		"var1": "val1",
		"var2": "val2",
		"var3": "val3",
	}
	config2 := map[string]string{
		"var4": "val4",
	}
	mockGetParametersByPathPagesWithContext(ssmMock, servicePath, config1, config2)

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	gotConfig, err := ps.ReadAll(ctx, servicePath)
	require.NoError(t, err)

	expectedConfig := map[string]string{}
	for key, value := range config1 {
		expectedConfig[key] = value
	}
	for key, value := range config2 {
		expectedConfig[key] = value
	}

	requireConfigEqual(t, expectedConfig, gotConfig)
}

// TestParameterStoreReadAllServiceNotExists verifies that ReadAll returns
// storage.ErrConfigNotFound when given a non-existing service name.
func TestParameterStoreReadAllServiceNotExists(t *testing.T) {
	const servicePath = "/service/name"
	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.MockGetParametersByPath = func(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
		return nil, ErrParameterNotFound
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	_, err := ps.ReadAll(ctx, servicePath)
	require.ErrorIs(t, err, storage.ErrConfigNotFound)
}

// TestParameterStoreReadKeysAllExist verifies that ReadKeys retrieves all
// requested keys when they all exist in AWS Parameter Store.
func TestParameterStoreReadKeysAllExist(t *testing.T) {
	const (
		servicePath = "/service/env"
	)

	const numKeys = 5
	expectedKeys := make([]string, numKeys)
	expectedConfig := make(map[string]string, numKeys)
	for i := 0; i < numKeys; i++ {
		expectedKeys[i] = fmt.Sprintf("var_%d", i)
		expectedConfig[expectedKeys[i]] = fmt.Sprintf("val %d", i)
	}

	ssmMock := &parameterstore.MockSSMClient{}
	parameters := makeParameters(servicePath, expectedConfig)
	ssmMock.MockGetParameters = func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
		return &ssm.GetParametersOutput{Parameters: parameters}, nil
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	gotValues, err := ps.ReadKeys(ctx, servicePath, expectedKeys)
	require.NoError(t, err)

	require.Equal(t, len(expectedConfig), len(gotValues))
	for key, value := range expectedConfig {
		require.Equal(t, value, gotValues[key])
	}
}

// TestParameterStoreReadKeysNoKeys verifies that ReadKeys returns an empty
// result set and no error when attempting to read zero keys.
func TestParameterStoreReadKeysNoKeys(t *testing.T) {
	const (
		servicePath = "/service/env"
	)

	ssmMock := &parameterstore.MockSSMClient{}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	gotValues, err := ps.ReadKeys(ctx, servicePath, []string{})
	require.NoError(t, err)
	require.Equal(t, 0, len(gotValues))
}

// TestParameterStoreReadKeysOneNotExist verifies that no values are returned
// when at least one parameter is invalid.
func TestParameterStoreReadKeysOneNotExist(t *testing.T) {
	const (
		servicePath    = "/service/env"
		nonExistingKey = "non-existing-key"
	)

	parameters := makeParameters(servicePath, map[string]string{"var1": "val1"})
	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.MockGetParameters = func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
		return &ssm.GetParametersOutput{
			Parameters:        parameters,
			InvalidParameters: []string{nonExistingKey},
		}, nil
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	gotValues, err := ps.ReadKeys(ctx, servicePath, []string{nonExistingKey})
	require.Equal(t, storage.ErrConfigNotFound, err)
	require.Equal(t, 0, len(gotValues))
}

// TestDeleteExpectedKey verifies that the expected key is requested for
// deletion.
func TestDeleteExpectedKey(t *testing.T) {
	const (
		servicePath = "/service/env"
		key         = "thekey"
	)
	keyPath := path.Join(servicePath, key)
	parameters := []string{keyPath}

	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.MockDeleteParameters = func(ctx context.Context, params *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
		return &ssm.DeleteParametersOutput{DeletedParameters: parameters}, nil
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	err := ps.Delete(ctx, servicePath, key)
	require.NoError(t, err)

	require.True(t, ssmMock.DeleteParametersCalled)
}

// TestDeleteKeysExpectedKeys verifies that the expected keys are requested
// for deletion.
func TestDeleteKeysExpectedKeys(t *testing.T) {
	const (
		servicePath = "/service/env"
		key1        = "first_key"
		key2        = "second_key"
	)
	key1Path := path.Join(servicePath, key1)
	key2Path := path.Join(servicePath, key2)
	parameters := []string{key1Path, key2Path}

	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.MockDeleteParameters = func(ctx context.Context, params *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
		return &ssm.DeleteParametersOutput{DeletedParameters: parameters}, nil
	}

	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	err := ps.DeleteKeys(ctx, servicePath, []string{key1, key2})
	require.NoError(t, err)
}

func requireConfigEqual(t *testing.T, expected, got map[string]string) {
	require.Equal(t, len(expected), len(got))
	for key, value := range expected {
		require.Equal(t, value, got[key])
	}
}

func mockGetParametersByPathPagesWithContext(ssmMock *parameterstore.MockSSMClient, servicePath string, configPages ...map[string]string) {
	batchI := 0
	ssmMock.MockGetParametersByPath = func(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
		output := &ssm.GetParametersByPathOutput{
			Parameters: makeParameters(servicePath, configPages[batchI]),
		}

		batchI += 1
		lastPage := batchI == len(configPages)
		if !lastPage {
			output.NextToken = aws.String("more data!")
		}

		return output, nil
	}
}

func makeParameters(servicePath string, config map[string]string) []types.Parameter {
	parameters := make([]types.Parameter, 0, len(config))

	for key, value := range config {
		parameters = append(parameters, types.Parameter{
			Name:  aws.String(path.Join(servicePath, key)),
			Value: aws.String(value),
		})
	}

	return parameters
}
