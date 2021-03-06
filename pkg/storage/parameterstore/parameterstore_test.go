package parameterstore_test

import (
	"context"
	"fmt"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/micvbang/confman-go/pkg/storage/parameterstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	parameters := makeParameters(servicePath, map[string]string{
		key: expectedValue,
	})
	ssmMock.On("GetParametersWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.GetParametersOutput{Parameters: parameters}, nil)

	log := logger.LogrusWrapper{logrus.New()}
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
	ssmMock.On("GetParametersWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return((*ssm.GetParametersOutput)(nil), &ssm.ParameterNotFound{})

	log := logger.LogrusWrapper{logrus.New()}
	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	_, err := ps.Read(ctx, "/some/service", "some-key")

	require.Equal(t, storage.ErrConfigNotFound, err)
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
	parameters := makeParameters(servicePath, map[string]string{key: "not value"})

	ssmMock.On("GetParametersWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.GetParametersOutput{Parameters: parameters}, nil)

	ssmMock.On("PutParameterWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.PutParameterOutput{}, nil)
	defer ssmMock.AssertExpectations(t)

	log := logger.LogrusWrapper{logrus.New()}
	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	err := ps.Write(ctx, servicePath, key, value)

	require.NoError(t, err)
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
	parameters := makeParameters(servicePath, map[string]string{key: value})

	ssmMock.On("GetParametersWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.GetParametersOutput{Parameters: parameters}, nil)

	defer ssmMock.AssertExpectations(t)

	log := logger.LogrusWrapper{logrus.New()}
	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	err := ps.Write(ctx, servicePath, key, value)

	require.NoError(t, err)
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

	ssmMock.On("GetParametersWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return((*ssm.GetParametersOutput)(nil), &ssm.ParameterNotFound{})

	ssmMock.On("PutParameterWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.PutParameterOutput{}, nil)
	defer ssmMock.AssertExpectations(t)

	log := logger.LogrusWrapper{logrus.New()}
	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	err := ps.Write(ctx, servicePath, key, value)

	require.NoError(t, err)
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

	mockGetParametersByPathPagesWithContext(ssmMock, servicePath, expectedConfig).Once()

	log := logger.LogrusWrapper{logrus.New()}
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
	mockGetParametersByPathPagesWithContext(ssmMock, servicePath, config1, config2).Once()

	defer ssmMock.AssertExpectations(t)

	log := logger.LogrusWrapper{logrus.New()}
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
	ssmMock.On("GetParametersByPathPagesWithContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.ParameterNotFound{})

	log := logger.LogrusWrapper{logrus.New()}
	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	_, err := ps.ReadAll(ctx, servicePath)
	require.Equal(t, storage.ErrConfigNotFound, err)
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
	ssmMock.On("GetParametersWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.GetParametersOutput{Parameters: parameters}, nil)

	log := logger.LogrusWrapper{logrus.New()}
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

	log := logger.LogrusWrapper{logrus.New()}
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
	ssmMock.On("GetParametersWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.GetParametersOutput{
			Parameters:        parameters,
			InvalidParameters: []*string{aws.String(nonExistingKey)},
		}, nil)

	log := logger.LogrusWrapper{logrus.New()}
	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	gotValues, err := ps.ReadKeys(ctx, servicePath, []string{nonExistingKey})
	require.Equal(t, storage.ErrConfigNotFound, err)
	require.Equal(t, 0, len(gotValues))
}

// TODO: test Delete* functionality

// TestDeleteExpectedKey verifies that the expected key is requested for
// deletion.
func TestDeleteExpectedKey(t *testing.T) {
	const (
		servicePath = "/service/env"
		key         = "thekey"
	)
	keyPath := path.Join(servicePath, key)
	parameters := []*string{&keyPath}

	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.On("DeleteParametersWithContext", mock.Anything, &ssm.DeleteParametersInput{Names: parameters}, mock.Anything).
		Return(&ssm.DeleteParametersOutput{DeletedParameters: parameters}, nil)

	log := logger.LogrusWrapper{logrus.New()}
	ps := parameterstore.New(log, ssmMock, "kms key id")

	ctx := context.Background()
	err := ps.Delete(ctx, servicePath, key)
	require.NoError(t, err)
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
	parameters := []*string{&key1Path, &key2Path}

	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.On("DeleteParametersWithContext", mock.Anything, &ssm.DeleteParametersInput{Names: parameters}, mock.Anything).
		Return(&ssm.DeleteParametersOutput{DeletedParameters: parameters}, nil)

	log := logger.LogrusWrapper{logrus.New()}
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

func mockGetParametersByPathPagesWithContext(ssmMock *parameterstore.MockSSMClient, servicePath string, configPages ...map[string]string) *mock.Call {
	return ssmMock.On("GetParametersByPathPagesWithContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			f := args.Get(2).(func(*ssm.GetParametersByPathOutput, bool) bool)

			for i, config := range configPages {
				output := &ssm.GetParametersByPathOutput{
					Parameters: makeParameters(servicePath, config),
				}

				lastPage := i == len(configPages)-1
				f(output, lastPage)
			}
		})
}

func makeParameters(servicePath string, config map[string]string) []*ssm.Parameter {
	parameters := make([]*ssm.Parameter, 0, len(config))

	for key, value := range config {
		parameters = append(parameters, &ssm.Parameter{
			Name:  aws.String(path.Join(servicePath, key)),
			Value: aws.String(value),
		})
	}

	return parameters
}
