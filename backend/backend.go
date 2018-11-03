package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/ssmvars"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
)

const (
	archivePrefix = "archive"
	livePrefix    = "live"
)

var defaultEntropySource io.Reader

// Backend is an S3 implementation of the secretservice backend.
type Backend struct {
	ssmvars.ReadWriter

	s3         s3iface.S3API
	bucketName *string
}

// New returns an implementation of Secret Service backend.
func New(ssm ssmvars.ReadWriter, s3 s3iface.S3API, bucketName string) *Backend {
	return &Backend{
		ReadWriter: ssm,
		s3:         s3,
		bucketName: aws.String(bucketName),
	}
}

// CreateRelease creates a release with a given set of variables.
func (b *Backend) CreateRelease(ctx context.Context, scopeName string, variables []*ssmvars.Variable) (*secretservice.Release, error) {
	ulid, err := ulid.New(ulid.Now(), defaultEntropySource)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate an ID")
	}

	scope, err := b.Scope(ctx, scopeName)
	if err != nil {
		return nil, err
	}

	release := &secretservice.Release{
		ID:        ulid.String(),
		ScopeName: scopeName,
		Live:      true,
		Variables: variables,
	}

	body, err := json.Marshal(release)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal the release")
	}

	kmsKeyID := aws.String(scope.KMSKeyID)
	archiveKey := b.objectKey(release.ScopeName, archivePrefix, release.ID)
	_, err = b.s3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:                 bytes.NewReader(body),
		Bucket:               b.bucketName,
		Key:                  archiveKey,
		SSEKMSKeyId:          kmsKeyID,
		ServerSideEncryption: aws.String("aws:kms"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not put archive object to S3")
	}

	_, err = b.s3.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:               b.bucketName,
		CopySource:           aws.String(path.Join(*b.bucketName, *archiveKey)),
		Key:                  b.objectKey(release.ScopeName, livePrefix, release.ID),
		SSEKMSKeyId:          kmsKeyID,
		ServerSideEncryption: aws.String("aws:kms"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not copy live version to S3")
	}

	return release, nil
}

// GetRelease retrieves a release given its ID.
func (b *Backend) GetRelease(ctx context.Context, scopeName, releaseID string) (*secretservice.Release, error) {
	ret, err := b.getRelease(ctx, livePrefix, scopeName, releaseID)
	if err == nil {
		return ret, nil
	}

	if aerr, ok := errors.Cause(err).(awserr.Error); !ok || aerr.Code() != "NotFound" {
		return nil, err
	}

	return b.getRelease(ctx, archivePrefix, scopeName, releaseID)
}

// ArchiveRelease archives a release.
func (b *Backend) ArchiveRelease(ctx context.Context, scopeName, releaseID string) error {
	_, err := b.s3.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: b.bucketName,
		Key:    b.objectKey(scopeName, "live", releaseID),
	})

	return errors.Wrap(err, "could not remove live object from S3")
}

// Scope returns scope by its name.
func (b *Backend) Scope(ctx context.Context, scopeName string) (*secretservice.Scope, error) {
	scopeVar, err := b.ShowVariable(ctx, "scopes", scopeName)
	if err != nil {
		return nil, errors.Wrapf(err, "could not find scope %s", scopeName)
	}
	return &secretservice.Scope{Name: scopeName, KMSKeyID: scopeVar.Value}, nil
}

func (b *Backend) getRelease(ctx context.Context, prefix, scopeName, releaseID string) (*secretservice.Release, error) {
	output, err := b.s3.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: b.bucketName,
		Key:    b.objectKey(scopeName, prefix, releaseID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve object from ")
	}
	defer output.Body.Close()

	release := new(secretservice.Release)
	if err := json.NewDecoder(output.Body).Decode(&release); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal release")
	}

	release.ID = releaseID
	release.Live = prefix == livePrefix
	release.ScopeName = scopeName

	return release, nil
}

func (b *Backend) objectKey(scopeName, prefix, releaseID string) *string {
	return aws.String(path.Join(scopeName, prefix, releaseID))
}
