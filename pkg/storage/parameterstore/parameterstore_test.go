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
		serviceName = "/service/env"
		key         = "key"
		value       = "the value"
	)

	ssmMock := &parameterstore.MockSSMClient{}
	parameter := &ssm.Parameter{
		Name:  aws.String(path.Join(serviceName, key)),
		Value: aws.String(value),
	}
	ssmMock.On("GetParameterWithContext", mock.Anything, mock.Anything).
		Return(&ssm.GetParameterOutput{Parameter: parameter}, nil)

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	v, err := ps.Read(ctx, key)

	require.NoError(t, err)
	require.Equal(t, value, v)
}

// TestParameterStoreReadNotExists verifies that the expected error is returned
// when the given key does not exist.
func TestParameterStoreReadNotExists(t *testing.T) {
	ssmMock := &parameterstore.MockSSMClient{}
	ssmMock.On("GetParameterWithContext", mock.Anything, mock.Anything).
		Return((*ssm.GetParameterOutput)(nil), &ssm.ParameterNotFound{})

	ps := parameterstore.New(ssmMock, "kms key id", "/some/service")

	ctx := context.Background()
	_, err := ps.Read(ctx, "some-key")

	require.Equal(t, storage.ErrConfigNotFound, err)
}

// TestParameterStorePutExistsNotEqual verifies that ParameterStore checks if
// the given key already exists and adds it if the given value is not equal to
// the existing one.
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

	ssmMock.On("GetParameterWithContext", mock.Anything, mock.Anything).
		Return(&ssm.GetParameterOutput{Parameter: parameter}, nil)

	ssmMock.On("PutParameterWithContext", mock.Anything, mock.Anything).
		Return(&ssm.PutParameterOutput{}, nil)
	defer ssmMock.AssertExpectations(t)

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	err := ps.Add(ctx, key, value)

	require.NoError(t, err)
}

// TestParameterStorePutExistsEqual verifies that ParameterStore checks if
// the given key already exists and does not add it, if the given key is equal
// to the existing one.
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

	ssmMock.On("GetParameterWithContext", mock.Anything, mock.Anything).
		Return(&ssm.GetParameterOutput{Parameter: parameter}, nil)

	defer ssmMock.AssertExpectations(t)

	ps := parameterstore.New(ssmMock, "kms key id", serviceName)

	ctx := context.Background()
	err := ps.Add(ctx, key, value)

	require.NoError(t, err)
}
