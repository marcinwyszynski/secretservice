package resolver

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/ssmvars"
	"github.com/pkg/errors"
)

type scopeResolver struct {
	backend secretservice.Backend
	wraps   *secretservice.Scope
}

// id: ID!
func (s *scopeResolver) ID() graphql.ID {
	return graphql.ID(s.wraps.Name)
}

type diffArgs struct {
	Since graphql.ID
}

// diff(since: ID!) Diff!
func (s *scopeResolver) Diff(ctx context.Context, args diffArgs) (*diffResolver, error) {
	newVariables, err := s.workspace(ctx)
	if err != nil {
		return nil, err
	}

	release, err := s.backend.GetRelease(ctx, s.wraps.Name, string(args.Since))
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve old release")
	}

	return newDiffResolver(release.Variables, newVariables), nil
}

// kmsKeyId: String!
func (s *scopeResolver) KMSKeyID() string {
	return s.wraps.KMSKeyID
}

type releaseArgs struct {
	ID graphql.ID
}

// release(id: ID!): Release!
func (s *scopeResolver) Release(args releaseArgs) *releaseResolver {
	return newReleaseResolver(s.backend, args.ID, s.wraps)
}

type releasesArgs struct {
	Before *graphql.ID
}

// releases(scopeId: ID!, before: ID): [Release!]!
func (s *scopeResolver) Releases(ctx context.Context, args releasesArgs) ([]*releaseResolver, error) {
	var before *string
	if args.Before != nil {
		before = aws.String(string(*args.Before))
	}

	ids, err := s.backend.ListReleases(ctx, s.wraps.Name, before)
	if err != nil {
		return nil, errors.Wrap(err, "could not list release IDs")
	}

	ret := make([]*releaseResolver, len(ids), len(ids))
	for index, id := range ids {
		ret[index] = newReleaseResolver(s.backend, graphql.ID(id), s.wraps)
	}

	return ret, nil
}

// variables: [Variable!]!
func (s *scopeResolver) Variables(ctx context.Context) ([]*variableResolver, error) {
	variables, err := s.workspace(ctx)
	if err != nil {
		return nil, err
	}

	num := len(variables)
	ret := make([]*variableResolver, num, num)
	for i, variable := range variables {
		ret[i] = &variableResolver{wraps: variable}
	}
	return ret, nil
}

func (s *scopeResolver) workspace(ctx context.Context) ([]*ssmvars.Variable, error) {
	ret, err := s.backend.ListVariables(ctx, fmt.Sprintf("workspace/%s", s.wraps.Name))
	if err != nil {
		return nil, errors.Wrap(err, "could not get workspace")
	}
	return ret, nil
}
