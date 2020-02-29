package parameterstore

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/mock"
)

type MockSSMClient struct {
	ssmiface.SSMAPI
	mock.Mock
}

func (_m *MockSSMClient) GetParameterWithContext(ctx context.Context, input *ssm.GetParameterInput, options ...request.Option) (*ssm.GetParameterOutput, error) {
	ret := _m.Called(ctx, input, options)

	r0 := ret.Get(0).(*ssm.GetParameterOutput)
	r1 := ret.Error(1)

	return r0, r1
}

func (_m *MockSSMClient) GetParametersWithContext(ctx context.Context, input *ssm.GetParametersInput, options ...request.Option) (*ssm.GetParametersOutput, error) {
	ret := _m.Called(ctx, input, options)

	r0 := ret.Get(0).(*ssm.GetParametersOutput)
	r1 := ret.Error(1)

	return r0, r1
}

func (_m *MockSSMClient) PutParameterWithContext(ctx context.Context, input *ssm.PutParameterInput, options ...request.Option) (*ssm.PutParameterOutput, error) {
	ret := _m.Called(ctx, input, options)

	r0 := ret.Get(0).(*ssm.PutParameterOutput)
	r1 := ret.Error(1)

	return r0, r1
}

func (_m *MockSSMClient) GetParametersByPathPagesWithContext(ctx context.Context, input *ssm.GetParametersByPathInput, fn func(*ssm.GetParametersByPathOutput, bool) bool, options ...request.Option) error {
	ret := _m.Called(ctx, input, fn, options)

	r0 := ret.Error(0)

	return r0
}
