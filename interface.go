package secretservice

import (
	"context"

	"github.com/marcinwyszynski/ssmvars"
)

// Backend is an abstract representation of the AWS backing services.
type Backend interface {
	ssmvars.ReadWriter

	CreateRelease(ctx context.Context, scopeName string, variables []*ssmvars.Variable) (*Release, error)
	GetRelease(ctx context.Context, scopeName, releaseID string) (*Release, error)
	ArchiveRelease(ctx context.Context, scopeName, releaseID string) error
}
