package backend

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/ssmvars"
)

// Backend is an S3 implementation of the secretservice backend.
type Backend struct {
	ssmvars.ReadWriter

	bucketName string
	s3         s3iface.S3API
}

// CreateRelease creates a release with a given bunch of variables.
func (b *Backend) CreateRelease(ctx context.Context, scopeName string, variables []*ssmvars.Variable) (*secretservice.Release, error) {
	panic("TODO")
}

// GetRelease retrieves a release given its ID.
func (b *Backend) GetRelease(ctx context.Context, scopeName, releaseID string) (*secretservice.Release, error) {
	panic("TODO")
}

// ArchiveRelease archives a release.
func (b *Backend) ArchiveRelease(ctx context.Context, scopeName, releaseID string) error {
	panic("TODO")
}
