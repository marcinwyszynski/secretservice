package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/graph-gophers/graphql-go"
	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/ssmvars"
	"github.com/stretchr/testify/suite"
)

type scopeResolverTestSuite struct {
	suite.Suite

	backend *mockBackend
	ctx     context.Context
	scope   *secretservice.Scope

	sut *scopeResolver
}

func (s *scopeResolverTestSuite) SetupTest() {
	s.backend = new(mockBackend)
	s.ctx = context.Background()
	s.scope = &secretservice.Scope{Name: "scopeName", KMSKeyID: "kmsKeyID"}
	s.sut = &scopeResolver{backend: s.backend, wraps: s.scope}
}

func (s *scopeResolverTestSuite) TestID() {
	s.EqualValues("scopeName", s.sut.ID())
}

func (s *scopeResolverTestSuite) TestDiff_OK() {
	oldVariable := &ssmvars.Variable{Name: "OLD"}
	newVariable := &ssmvars.Variable{Name: "NEW"}

	s.backend.
		On("ListVariables", s.ctx, "workspace/scopeName").
		Return([]*ssmvars.Variable{newVariable}, nil)

	s.backend.
		On("GetRelease", s.ctx, "scopeName", "since").
		Return(&secretservice.Release{Variables: []*ssmvars.Variable{oldVariable}}, nil)

	diff, err := s.sut.Diff(s.ctx, diffArgs{Since: "since"})

	s.NoError(err)

	added := diff.Added()
	s.Len(added, 1)
	s.Equal(newVariable, added[0].wraps)

	deleted := diff.Deleted()
	s.Len(deleted, 1)
	s.Equal(oldVariable, deleted[0].wraps)
}

func (s *scopeResolverTestSuite) TestDiff_SSMFailure() {
	s.backend.
		On("ListVariables", s.ctx, "workspace/scopeName").
		Return([]*ssmvars.Variable(nil), errors.New("bacon"))

	diff, err := s.sut.Diff(s.ctx, diffArgs{Since: "since"})

	s.Nil(diff)
	s.EqualError(err, "could not get workspace: bacon")
}

func (s *scopeResolverTestSuite) TestDiff_S3Failure() {
	newVariable := &ssmvars.Variable{Name: "NEW"}

	s.backend.
		On("ListVariables", s.ctx, "workspace/scopeName").
		Return([]*ssmvars.Variable{newVariable}, nil)

	s.backend.
		On("GetRelease", s.ctx, "scopeName", "since").
		Return((*secretservice.Release)(nil), errors.New("bacon"))

	diff, err := s.sut.Diff(s.ctx, diffArgs{Since: "since"})

	s.Nil(diff)
	s.EqualError(err, "could not retrieve old release: bacon")
}

func (s *scopeResolverTestSuite) TestKMSKeyID() {
	s.Equal("kmsKeyID", s.sut.KMSKeyID())
}

func (s *scopeResolverTestSuite) TestRelease() {
	release := s.sut.Release(releaseArgs{ID: "releaseID"})

	s.EqualValues("releaseID", release.ID())
	s.Equal(s.scope, release.scope)
	s.Equal(s.backend, release.backend)
}

func (s *scopeResolverTestSuite) TestReleases_OK() {
	s.backend.
		On("ListReleases", s.ctx, "scopeName", aws.String("before")).
		Return([]string{"releaseID"}, nil)

	beforeArg := graphql.ID("before")
	ret, err := s.sut.Releases(s.ctx, releasesArgs{Before: &beforeArg})

	s.NoError(err)
	s.Len(ret, 1)

	release := ret[0]
	s.EqualValues("releaseID", release.ID())
	s.Equal(s.scope, release.scope)
	s.Equal(s.backend, release.backend)
}

func (s *scopeResolverTestSuite) TestReleases_BackendFailure() {
	s.backend.
		On("ListReleases", s.ctx, "scopeName", aws.String("before")).
		Return([]string(nil), errors.New("bacon"))

	beforeArg := graphql.ID("before")
	ret, err := s.sut.Releases(s.ctx, releasesArgs{Before: &beforeArg})

	s.Nil(ret)
	s.EqualError(err, "could not list release IDs: bacon")
}

func (s *scopeResolverTestSuite) TestVariables_OK() {
	variable := &ssmvars.Variable{Name: "NEW"}

	s.backend.
		On("ListVariables", s.ctx, "workspace/scopeName").
		Return([]*ssmvars.Variable{variable}, nil)

	ret, err := s.sut.Variables(s.ctx)

	s.NoError(err)
	s.Len(ret, 1)
	s.Equal(variable, ret[0].wraps)
}

func (s *scopeResolverTestSuite) TestVariables_BackendFailure() {
	s.backend.
		On("ListVariables", s.ctx, "workspace/scopeName").
		Return([]*ssmvars.Variable(nil), errors.New("bacon"))

	ret, err := s.sut.Variables(s.ctx)

	s.Nil(ret)
	s.EqualError(err, "could not get workspace: bacon")
}

func TestScopeResolver(t *testing.T) {
	suite.Run(t, new(scopeResolverTestSuite))
}
