package parameterstore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type MockSSMClient struct {
	PutParameterCalled bool
	MockPutParameter   func(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)

	GetParametersCalled bool
	MockGetParameters   func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error)

	GetParametersByPathCalled bool
	MockGetParametersByPath   func(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)

	DescribeParametersCalled bool
	MockDescribeParameters   func(ctx context.Context, params *ssm.DescribeParametersInput, optFns ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error)

	DeleteParametersCalled bool
	MockDeleteParameters   func(ctx context.Context, params *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error)
}

func (m *MockSSMClient) PutParameter(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
	m.PutParameterCalled = true
	return m.MockPutParameter(ctx, params, optFns...)
}
func (m *MockSSMClient) GetParameters(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
	m.GetParametersCalled = true
	return m.MockGetParameters(ctx, params, optFns...)
}
func (m *MockSSMClient) GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	m.GetParametersByPathCalled = true
	return m.MockGetParametersByPath(ctx, params, optFns...)
}
func (m *MockSSMClient) DescribeParameters(ctx context.Context, params *ssm.DescribeParametersInput, optFns ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error) {
	m.DescribeParametersCalled = true
	return m.MockDescribeParameters(ctx, params, optFns...)
}
func (m *MockSSMClient) DeleteParameters(ctx context.Context, params *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
	m.DeleteParametersCalled = true
	return m.MockDeleteParameters(ctx, params, optFns...)
}
