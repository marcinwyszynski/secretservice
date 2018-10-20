package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

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

	bucketName string
	s3         s3iface.S3API
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

	kmsKeyID := scope.KMSKeyID
	if err := b.putRelease(ctx, archivePrefix, kmsKeyID, release); err != nil {
		return nil, errors.Wrap(err, "could not put release in the archive")
	}
	if err := b.putRelease(ctx, livePrefix, kmsKeyID, release); err != nil {
		return nil, errors.Wrap(err, "could not make the release live")
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
		Bucket: aws.String(b.bucketName),
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
		Bucket: aws.String(b.bucketName),
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

func (b *Backend) putRelease(ctx context.Context, prefix, kmsKeyID string, release *secretservice.Release) error {
	body, err := json.Marshal(release)
	if err != nil {
		return errors.Wrap(err, "could not marshal the release")
	}

	_, err = b.s3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:                 bytes.NewReader(body),
		Bucket:               aws.String(b.bucketName),
		Key:                  b.objectKey(release.ScopeName, prefix, release.ID),
		SSEKMSKeyId:          aws.String(kmsKeyID),
		ServerSideEncryption: aws.String("aws:kms"),
	})

	return errors.Wrap(err, "could not put object in S3")
}

func (b *Backend) objectKey(scopeName, prefix, releaseID string) *string {
	return aws.String(fmt.Sprintf("%s/%s/%s", scopeName, prefix, releaseID))
}
