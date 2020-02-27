package parameterstore_test

import (
	"context"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/micvbang/confman-go/pkg/storage/parameterstore"
)

func TestParameterStoreRead(t *testing.T) {
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
