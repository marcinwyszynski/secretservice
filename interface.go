package secretservice

import (
	"context"

	"github.com/marcinwyszynski/ssmvars"
)

// Backend is an abstract representation of the AWS backing services.
type Backend interface {
	ssmvars.ReadWriter

	ArchiveRelease(ctx context.Context, scopeName, releaseID string) error
	CreateRelease(ctx context.Context, scopeName string, variables []*ssmvars.Variable) (*Release, error)
	GetRelease(ctx context.Context, scopeName, releaseID string) (*Release, error)
	ListReleases(ctx context.Context, scopeName string, before *string) ([]string, error)
	Scope(ctx context.Context, scopeName string) (*Scope, error)
}
