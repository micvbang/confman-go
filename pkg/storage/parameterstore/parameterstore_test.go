package parameterstore_test

import (
	"context"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gitlab.com/micvbang/confman-go/pkg/storage/parameterstore"
)

// TestParameterStoreReadExists verifies that ParameterStore returns the
// expected value when the key to be read exists in AWS Parameter Store.
func TestParameterStoreReadExists(t *testing.T) {
	const (
		serviceName   = "/service/env"
		key           = "key"
		expectedValue = "the value"
	)

	ssmMock := &parameterstore.MockSSMClient{}
	parameter := &ssm.Parameter{
		Name:  aws.String(path.Join(serviceName, key)),
		Value: aws.String(expectedValue),
	}
	ssmMock.On("GetParameterWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.GetParameterOutput{Parameter: parameter}, nil)

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	gotValue, err := ps.Read(ctx, key)

	require.NoError(t, err)
	require.Equal(t, expectedValue, gotValue)
}

// TestParameterStoreReadNotExists verifies that ErrConfigNotFound is returned
// when the given key does not exist.
func TestParameterStoreReadNotExists(t *testing.T) {
	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.On("GetParameterWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return((*ssm.GetParameterOutput)(nil), &ssm.ParameterNotFound{})

	ps := parameterstore.New(ssmMock, "kms key id", "/some/service")

	ctx := context.Background()
	_, err := ps.Read(ctx, "some-key")

	require.Equal(t, storage.ErrConfigNotFound, err)
}

// TestParameterStorePutExistsNotEqual verifies that Put updates the given key
// in AWS Parameter Store when the key already exists, but the stored value is
// not equal to the input.
func TestParameterStorePutExistsNotEqual(t *testing.T) {
	const (
		serviceName = "/service/env"
		key         = "key"
		value       = "the value"
	)

	ssmMock := &parameterstore.MockSSMClient{}
	parameter := &ssm.Parameter{
		Name:  aws.String(path.Join(serviceName, key)),
		Value: aws.String("not value"),
	}

	ssmMock.On("GetParameterWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.GetParameterOutput{Parameter: parameter}, nil)

	ssmMock.On("PutParameterWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.PutParameterOutput{}, nil)
	defer ssmMock.AssertExpectations(t)

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	err := ps.Add(ctx, key, value)

	require.NoError(t, err)
}

// TestParameterStorePutExistsEqual verifies that Put does not update the
// given key in AWS Parameter Store when the key/value pair already exists,
// and the stored value is equal to the input.
func TestParameterStorePutExistsEqual(t *testing.T) {
	const (
		serviceName = "/service/env"
		key         = "key"
		value       = "the value"
	)

	ssmMock := &parameterstore.MockSSMClient{}
	parameter := &ssm.Parameter{
		Name:  aws.String(path.Join(serviceName, key)),
		Value: aws.String(value),
	}

	ssmMock.On("GetParameterWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.GetParameterOutput{Parameter: parameter}, nil)

	defer ssmMock.AssertExpectations(t)

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	err := ps.Add(ctx, key, value)

	require.NoError(t, err)
}

// TestParameterStorePutNotExists verifies that Put adds the given
// key/value pair to AWS Parameter Store when the given key does not already
// exist.
func TestParameterStorePutNotExists(t *testing.T) {
	const (
		serviceName = "/service/env"
		key         = "key"
		value       = "the value"
	)

	ssmMock := &parameterstore.MockSSMClient{}

	ssmMock.On("GetParameterWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return((*ssm.GetParameterOutput)(nil), &ssm.ParameterNotFound{})

	ssmMock.On("PutParameterWithContext", mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.PutParameterOutput{}, nil)
	defer ssmMock.AssertExpectations(t)

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	err := ps.Add(ctx, key, value)

	require.NoError(t, err)
}

// TestParameterStoreReadAllServiceExists verifies that configs are retrieved
// correctly for services with a single page worth of AWS Parameter Store
// parameters.
func TestParameterStoreReadAllServiceExists(t *testing.T) {
	const serviceName = "/service-name"

	ssmMock := &parameterstore.MockSSMClient{}

	expectedConfig := map[string]string{
		"var1": "val1",
		"var2": "val2",
		"var3": "val3",
	}

	mockGetParametersByPathPagesWithContext(ssmMock, serviceName, expectedConfig).Once()

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	gotConfig, err := ps.ReadAll(ctx)
	require.NoError(t, err)

	requireConfigEqual(t, expectedConfig, gotConfig)
}

// TestParameterStoreReadAllServiceExistsMultiplePages verifies that configs
// are retrieved corrcetly for services with multiple AWS Parameter Store pages
// worth of parameters.
func TestParameterStoreReadAllServiceExistsMultiplePages(t *testing.T) {
	const serviceName = "/service-name"

	ssmMock := &parameterstore.MockSSMClient{}

	config1 := map[string]string{
		"var1": "val1",
		"var2": "val2",
		"var3": "val3",
	}
	config2 := map[string]string{
		"var4": "val4",
	}
	mockGetParametersByPathPagesWithContext(ssmMock, serviceName, config1, config2).Once()

	defer ssmMock.AssertExpectations(t)

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	gotConfig, err := ps.ReadAll(ctx)
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
	const serviceName = "/service/name"
	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.On("GetParametersByPathPagesWithContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&ssm.ParameterNotFound{})

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	_, err := ps.ReadAll(ctx)
	require.Equal(t, storage.ErrConfigNotFound, err)
}

func requireConfigEqual(t *testing.T, expected, got map[string]string) {
	require.Equal(t, len(expected), len(got))
	for key, value := range expected {
		require.Equal(t, value, got[key])
	}
}

func mockGetParametersByPathPagesWithContext(ssmMock *parameterstore.MockSSMClient, serviceName string, configPages ...map[string]string) *mock.Call {
	return ssmMock.On("GetParametersByPathPagesWithContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			f := args.Get(2).(func(*ssm.GetParametersByPathOutput, bool) bool)

			for i, config := range configPages {
				output := &ssm.GetParametersByPathOutput{
					Parameters: makeParameters(serviceName, config),
				}

				lastPage := i == len(configPages)-1
				f(output, lastPage)
			}
		})
}

func makeParameters(serviceName string, config map[string]string) []*ssm.Parameter {
	parameters := make([]*ssm.Parameter, 0, len(config))

	for key, value := range config {
		parameters = append(parameters, &ssm.Parameter{
			Name:  aws.String(path.Join(serviceName, key)),
			Value: aws.String(value),
		})
	}

	return parameters
}
