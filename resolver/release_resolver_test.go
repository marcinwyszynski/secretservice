package resolver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/ssmvars"
	"github.com/oklog/ulid"
	"github.com/stretchr/testify/suite"
)

type releaseResolverTestSuite struct {
	suite.Suite

	backend *mockBackend
	ctx     context.Context

	sut *releaseResolver
}

func (r *releaseResolverTestSuite) SetupTest() {
	r.backend = new(mockBackend)
	r.ctx = context.Background()
	r.sut = newReleaseResolver(r.backend, "releaseID", &secretservice.Scope{
		Name:     "scopeName",
		KMSKeyID: "kmsKeyID",
	})
}

func (r *releaseResolverTestSuite) TestID() {
	r.EqualValues("releaseID", r.sut.ID())
}

func (r *releaseResolverTestSuite) TestDiff_OK() {
	oldVariable := &ssmvars.Variable{Name: "OLD"}
	newVariable := &ssmvars.Variable{Name: "NEW"}

	r.backend.
		On("GetRelease", r.ctx, "scopeName", "releaseID").
		Return(&secretservice.Release{Variables: []*ssmvars.Variable{newVariable}}, nil)

	r.backend.
		On("GetRelease", r.ctx, "scopeName", "compareID").
		Return(&secretservice.Release{Variables: []*ssmvars.Variable{oldVariable}}, nil)

	diff, err := r.sut.Diff(r.ctx, diffArgs{Since: "compareID"})

	r.NoError(err)

	added := diff.Added()
	r.Len(added, 1)
	r.Equal(newVariable, added[0].wraps)

	r.Empty(diff.Changed())

	deleted := diff.Deleted()
	r.Len(deleted, 1)
	r.Equal(oldVariable, deleted[0].wraps)
}

func (r *releaseResolverTestSuite) TestDiff_BackendFailureBase() {
	r.backend.
		On("GetRelease", r.ctx, "scopeName", "releaseID").
		Return((*secretservice.Release)(nil), errors.New("bacon"))

	diff, err := r.sut.Diff(r.ctx, diffArgs{Since: "compareID"})

	r.Nil(diff)
	r.EqualError(err, "could not lazily retrieve release: bacon")
}

func (r *releaseResolverTestSuite) TestDiff_BackendFailureCompare() {
	r.backend.
		On("GetRelease", r.ctx, "scopeName", "releaseID").
		Return(new(secretservice.Release), nil)

	r.backend.
		On("GetRelease", r.ctx, "scopeName", "compareID").
		Return((*secretservice.Release)(nil), errors.New("bacon"))

	diff, err := r.sut.Diff(r.ctx, diffArgs{Since: "compareID"})

	r.Nil(diff)
	r.EqualError(err, "could not pull old release: bacon")
}

func (r *releaseResolverTestSuite) TestLive_Live() {
	r.backend.
		On("GetRelease", r.ctx, "scopeName", "releaseID").
		Return(&secretservice.Release{Live: true}, nil)

	ret, err := r.sut.Live(r.ctx)

	r.NoError(err)
	r.True(ret)
}

func (r *releaseResolverTestSuite) TestLive_NotLive() {
	r.backend.
		On("GetRelease", r.ctx, "scopeName", "releaseID").
		Return(&secretservice.Release{Live: false}, nil)

	ret, err := r.sut.Live(r.ctx)

	r.NoError(err)
	r.False(ret)
}

func (r *releaseResolverTestSuite) TestLive_BackendFailure() {
	r.backend.
		On("GetRelease", r.ctx, "scopeName", "releaseID").
		Return((*secretservice.Release)(nil), errors.New("bacon"))

	ret, err := r.sut.Live(r.ctx)

	r.EqualError(err, "could not lazily retrieve release: bacon")
	r.False(ret)
}

func (r *releaseResolverTestSuite) TestTimestamp_OK() {
	r.sut.wraps = &secretservice.Release{
		ID: ulid.MustNew(ulid.MaxTime()-ulid.Now(), nil).String(),
	}

	timestamp, err := r.sut.Timestamp()

	r.NoError(err)
	r.InDelta(time.Now().Unix(), timestamp, 1)
}

func (r *releaseResolverTestSuite) TestTimestamp_FailedToParse() {
	r.sut.wraps = &secretservice.Release{ID: "bacon"}

	timestamp, err := r.sut.Timestamp()

	r.EqualError(err, "could not parse release ID as ULID: ulid: bad data size when unmarshaling")
	r.EqualValues(-1, timestamp)
}

func (r *releaseResolverTestSuite) TestVariables_OK() {
	variable := &ssmvars.Variable{Name: "BACON"}

	r.backend.
		On("GetRelease", r.ctx, "scopeName", "releaseID").
		Return(&secretservice.Release{Variables: []*ssmvars.Variable{variable}}, nil)

	variables, err := r.sut.Variables(r.ctx)

	r.NoError(err)
	r.Len(variables, 1)
	r.Equal(variable, variables[0].wraps)
}

func (r *releaseResolverTestSuite) TestVariables_BackendFailure() {
	r.backend.
		On("GetRelease", r.ctx, "scopeName", "releaseID").
		Return((*secretservice.Release)(nil), errors.New("bacon"))

	ret, err := r.sut.Variables(r.ctx)

	r.Nil(ret)
	r.EqualError(err, "could not lazily retrieve release: bacon")
}

func TestReleaseResolver(t *testing.T) {
	suite.Run(t, new(releaseResolverTestSuite))
}
