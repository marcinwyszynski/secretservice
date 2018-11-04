package backend_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/mock"
)

type mockS3 struct {
	mock.Mock
	s3iface.S3API
}

func (m *mockS3) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *mockS3) CopyObjectWithContext(ctx aws.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.CopyObjectOutput), args.Error(1)
}

func (m *mockS3) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *mockS3) DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}

func (m *mockS3) ListObjectsV2WithContext(ctx aws.Context, input *s3.ListObjectsV2Input, opts ...request.Option) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}
