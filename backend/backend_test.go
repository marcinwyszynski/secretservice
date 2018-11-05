package backend_test

import (
	"context"
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/marcinwyszynski/secretservice/backend"
	"github.com/marcinwyszynski/ssmvars"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	bucketName = "bucketName"
	kmsKeyID   = "kmsKeyID"
	releaseID  = "releaseID"
	scopeName  = "scopeName"
)

var variables = []*ssmvars.Variable{{
	Name:  "bacon",
	Value: "tasty",
}}

type backendTestSuite struct {
	suite.Suite

	ctx     context.Context
	ssmvars *mockSSMVars
	s3      *mockS3

	sut *backend.Backend
}

func (b *backendTestSuite) SetupTest() {
	b.ctx = context.Background()
	b.ssmvars = new(mockSSMVars)
	b.s3 = new(mockS3)
	b.sut = backend.New(b.ssmvars, b.s3, bucketName)
}

func (b *backendTestSuite) TestCreateRelease_OK() {
	b.withShowVariable(&ssmvars.Variable{Name: scopeName, Value: kmsKeyID}, nil)
	b.withPutObject(nil)
	b.withCopyObject(nil)

	release, err := b.sut.CreateRelease(b.ctx, scopeName, variables)

	b.NoError(err)
	b.NotEmpty(release.ID)

	timestamp, err := release.Timestamp()
	b.NoError(err)
	b.InDelta(timestamp, time.Now().Unix(), 1)

	b.True(release.Live)
	b.Equal(scopeName, release.ScopeName)
}

func (b *backendTestSuite) TestCreateRelease_FailScope() {
	b.withShowVariable(nil, errors.New("bacon"))

	release, err := b.sut.CreateRelease(b.ctx, scopeName, variables)

	b.Nil(release)
	b.EqualError(err, `could not find scope "scopeName": bacon`)
}

func (b *backendTestSuite) TestCreateRelease_FailPut() {
	b.withShowVariable(&ssmvars.Variable{Name: scopeName, Value: kmsKeyID}, nil)
	b.withPutObject(errors.New("bacon"))

	release, err := b.sut.CreateRelease(b.ctx, scopeName, variables)

	b.Nil(release)
	b.EqualError(err, "could not put archive object to S3: bacon")
}

func (b *backendTestSuite) TestCreateRelease_FailCopy() {
	b.withShowVariable(&ssmvars.Variable{Name: scopeName, Value: kmsKeyID}, nil)
	b.withPutObject(nil)
	b.withCopyObject(errors.New("bacon"))

	release, err := b.sut.CreateRelease(b.ctx, scopeName, variables)

	b.Nil(release)
	b.EqualError(err, "could not copy live version on S3: bacon")
}

func (b *backendTestSuite) TestGetRelease_OK() {
	b.withGetObject(`{"variables":[{"Name":"bacon","Value":"tasty","WriteOnly":true}]}`, nil)
	b.withLiveObjects(nil, "scopeName/live/releaseID")

	release, err := b.sut.GetRelease(b.ctx, scopeName, releaseID)

	b.NotNil(release)
	b.NoError(err)
	b.Equal(releaseID, release.ID)
	b.True(release.Live)
	b.Len(release.Variables, 1)

	variable := release.Variables[0]
	b.Equal("bacon", variable.Name)
	b.Equal("tasty", variable.Value)
	b.True(variable.WriteOnly)
}

func (b *backendTestSuite) TestGetRelease_NotLive() {
	b.withGetObject(`{"variables":[{"Name":"bacon","Value":"tasty","WriteOnly":true}]}`, nil)
	b.withLiveObjects(nil)

	release, err := b.sut.GetRelease(b.ctx, scopeName, releaseID)

	b.NoError(err)
	b.False(release.Live)
}

func (b *backendTestSuite) TestGetRelease_FailGet() {
	b.withGetObject("", errors.New("bacon"))

	release, err := b.sut.GetRelease(b.ctx, scopeName, releaseID)

	b.Nil(release)
	b.EqualError(err, "could not retrieve object from S3: bacon")
}

func (b *backendTestSuite) TestGetRelease_FailDecode() {
	b.withGetObject("invalid", nil)

	release, err := b.sut.GetRelease(b.ctx, scopeName, releaseID)

	b.Nil(release)
	b.EqualError(err, "could not unmarshal release: invalid character 'i' looking for beginning of value")
}

func (b *backendTestSuite) TestGetRelease_FailCheckLive() {
	b.withGetObject(`{"variables":[{"Name":"bacon","Value":"tasty","WriteOnly":true}]}`, nil)
	b.withLiveObjects(errors.New("bacon"))

	release, err := b.sut.GetRelease(b.ctx, scopeName, releaseID)

	b.Nil(release)
	b.EqualError(err, "could not check for live version presence: bacon")
}

func (b *backendTestSuite) TestArchiveRelease_OK() {
	b.withDeleteObject(nil)

	b.NoError(b.sut.ArchiveRelease(b.ctx, scopeName, releaseID))
}

func (b *backendTestSuite) TestArchiveRelease_FailDelete() {
	b.withDeleteObject(errors.New("bacon"))

	b.EqualError(
		b.sut.ArchiveRelease(b.ctx, scopeName, releaseID),
		"could not remove live object from S3: bacon",
	)
}

func (b *backendTestSuite) TestListReleases_OK() {
	b.withList(nil, nil, "scopeName/archive/bacon")

	releases, err := b.sut.ListReleases(b.ctx, scopeName, nil)

	b.NoError(err)
	b.Len(releases, 1)
	b.Equal("bacon", releases[0])
}

func (b *backendTestSuite) TestListReleases_Before() {
	b.withList(nil, nil, "scopeName/archive/bacon")

	releases, err := b.sut.ListReleases(b.ctx, scopeName, aws.String("before"))

	b.NoError(err)
	b.Len(releases, 1)
}

func (b *backendTestSuite) TestListReleases_Failure() {
	b.withList(errors.New("bacon"), nil)

	releases, err := b.sut.ListReleases(b.ctx, scopeName, aws.String("before"))

	b.Nil(releases)
	b.EqualError(err, "could not list objects with a prefix: bacon")
}

func (b *backendTestSuite) TestScope_OK() {
	b.withShowVariable(&ssmvars.Variable{Name: scopeName, Value: kmsKeyID}, nil)

	ret, err := b.sut.Scope(b.ctx, scopeName)
	b.NoError(err)

	b.Equal(scopeName, ret.Name)
	b.Equal(kmsKeyID, ret.KMSKeyID)
}

func (b *backendTestSuite) TestScope_FailShowVariable() {
	b.ssmvars.
		On("ShowVariable", b.ctx, "scopes", scopeName).
		Return((*ssmvars.Variable)(nil), errors.New("bacon"))

	ret, err := b.sut.Scope(b.ctx, scopeName)
	b.EqualError(err, `could not find scope "scopeName": bacon`)
	b.Nil(ret)
}

func (b *backendTestSuite) withCopyObject(err error) {
	b.s3.On(
		"CopyObjectWithContext",
		b.ctx,
		mock.MatchedBy(func(arg interface{}) bool {
			input, ok := arg.(*s3.CopyObjectInput)
			b.True(ok)

			b.Equal(bucketName, *input.Bucket)
			b.Contains(*input.CopySource, "bucketName/scopeName/archive/")
			b.Contains(*input.Key, "scopeName/live/")
			b.Equal(kmsKeyID, *input.SSEKMSKeyId)
			b.Equal("aws:kms", *input.ServerSideEncryption)

			return true
		}),
		[]request.Option(nil),
	).Return((*s3.CopyObjectOutput)(nil), err)
}

func (b *backendTestSuite) withDeleteObject(err error) {
	b.s3.On(
		"DeleteObjectWithContext",
		b.ctx,
		mock.MatchedBy(func(arg interface{}) bool {
			input, ok := arg.(*s3.DeleteObjectInput)
			b.True(ok)

			b.Equal(bucketName, *input.Bucket)
			b.Equal(*input.Key, "scopeName/live/releaseID")

			return true
		}),
		[]request.Option(nil),
	).Return((*s3.DeleteObjectOutput)(nil), err)
}

func (b *backendTestSuite) withGetObject(body string, err error) {
	b.s3.On(
		"GetObjectWithContext",
		b.ctx,
		mock.MatchedBy(func(arg interface{}) bool {
			input, ok := arg.(*s3.GetObjectInput)
			b.True(ok)

			b.Equal(bucketName, *input.Bucket)
			b.Equal("scopeName/archive/releaseID", *input.Key)

			return true
		}),
		[]request.Option(nil),
	).Return(&s3.GetObjectOutput{Body: ioutil.NopCloser(strings.NewReader(body))}, err)
}

func (b *backendTestSuite) withList(err error, before *string, keys ...string) {
	objects := make([]*s3.Object, len(keys), len(keys))
	for index, key := range keys {
		objects[index] = &s3.Object{Key: aws.String(key)}
	}

	b.s3.On(
		"ListObjectsV2WithContext",
		b.ctx,
		mock.MatchedBy(func(arg interface{}) bool {
			input, ok := arg.(*s3.ListObjectsV2Input)
			b.True(ok)

			b.Equal(bucketName, *input.Bucket)
			b.Equal("scopeName/archive/", *input.Prefix)
			b.EqualValues(10, *input.MaxKeys)

			if before != nil {
				b.Equal(*before, *input.StartAfter)
			}

			return true
		}),
		[]request.Option(nil),
	).Return(&s3.ListObjectsV2Output{Contents: objects}, err)
}

func (b *backendTestSuite) withLiveObjects(err error, keys ...string) {
	objects := make([]*s3.Object, len(keys), len(keys))
	for index, key := range keys {
		objects[index] = &s3.Object{Key: aws.String(key)}
	}

	b.s3.On(
		"ListObjectsV2WithContext",
		b.ctx,
		mock.MatchedBy(func(arg interface{}) bool {
			input, ok := arg.(*s3.ListObjectsV2Input)
			b.True(ok)

			b.Equal(bucketName, *input.Bucket)
			b.Equal(*input.Prefix, "scopeName/live/releaseID")

			return true
		}),
		[]request.Option(nil),
	).Return(&s3.ListObjectsV2Output{Contents: objects}, err)
}

func (b *backendTestSuite) withPutObject(err error) {
	b.s3.On(
		"PutObjectWithContext",
		b.ctx,
		mock.MatchedBy(func(arg interface{}) bool {
			input, ok := arg.(*s3.PutObjectInput)
			b.True(ok)

			data, err := ioutil.ReadAll(input.Body)
			b.NoError(err)
			b.Contains(string(data), "bacon")

			b.Equal(bucketName, *input.Bucket)
			b.Contains(*input.Key, "scopeName/archive/")
			b.Equal(kmsKeyID, *input.SSEKMSKeyId)
			b.Equal("aws:kms", *input.ServerSideEncryption)

			return true
		}),
		[]request.Option(nil),
	).Return((*s3.PutObjectOutput)(nil), err)
}

func (b *backendTestSuite) withShowVariable(ret *ssmvars.Variable, err error) {
	b.ssmvars.On("ShowVariable", b.ctx, "scopes", scopeName).Return(ret, err)
}

func TestBackend(t *testing.T) {
	suite.Run(t, new(backendTestSuite))
}
