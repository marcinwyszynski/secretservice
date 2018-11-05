package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/ssmvars"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type rootResolverTestSuite struct {
	suite.Suite

	backend *mockBackend
	ctx     context.Context
	sut     *rootResolver
}

func (r *rootResolverTestSuite) SetupTest() {
	r.backend = new(mockBackend)
	r.ctx = context.Background()
	r.sut = New(r.backend).(*rootResolver)
}

func (r *rootResolverTestSuite) TestScope_OK() {
	r.withScope(nil)

	ret, err := r.sut.Scope(r.ctx, scopeArgs{ScopeID: "scopeName"})

	r.NoError(err)
	r.EqualValues("scopeName", ret.ID())
	r.Equal(r.backend, ret.backend)
}

func (r *rootResolverTestSuite) TestScope_BackendFailure() {
	r.withScope(errors.New("bacon"))

	ret, err := r.sut.Scope(r.ctx, scopeArgs{ScopeID: "scopeName"})

	r.Nil(ret)
	r.EqualError(err, "could not retrieve scope: bacon")
}

func (r *rootResolverTestSuite) TestCreateScope_OK() {
	r.withListVariables("scopes", nil)
	r.withCreateVariable("scopes", &ssmvars.Variable{Name: "scopeName", Value: "kmsKeyID"}, nil)

	ret, err := r.sut.CreateScope(r.ctx, createScopeArgs{
		Name:     "scopeName",
		KMSKeyID: "kmsKeyID",
	})

	r.NoError(err)
	r.EqualValues("scopeName", ret.ID())
	r.Equal("kmsKeyID", ret.KMSKeyID())
	r.Equal(r.backend, ret.backend)
}

func (r *rootResolverTestSuite) TestCreateScope_AlreadyExists() {
	r.withListVariables("scopes", nil, &ssmvars.Variable{Name: "scopeName"})

	ret, err := r.sut.CreateScope(r.ctx, createScopeArgs{
		Name:     "scopeName",
		KMSKeyID: "kmsKeyID",
	})

	r.Nil(ret)
	r.EqualError(err, `scope "scopeName" already exists`)
}

func (r *rootResolverTestSuite) TestCreateScope_BackendFailure() {
	r.withListVariables("scopes", nil)
	r.withCreateVariable(
		"scopes",
		&ssmvars.Variable{Name: "scopeName", Value: "kmsKeyID"},
		errors.New("bacon"),
	)

	ret, err := r.sut.CreateScope(r.ctx, createScopeArgs{
		Name:     "scopeName",
		KMSKeyID: "kmsKeyID",
	})

	r.Nil(ret)
	r.EqualError(err, "could not create scope: bacon")
}

func (r *rootResolverTestSuite) TestAddVariable_OK() {
	r.withScope(nil)

	r.withCreateVariable(
		"workspace/scopeName",
		&ssmvars.Variable{Name: "name", Value: "value", WriteOnly: true},
		nil,
	)

	ret, err := r.addVariable()

	r.NoError(err)

	r.EqualValues("name", ret.ID())
	r.Nil(ret.Value())
}

func (r *rootResolverTestSuite) TestAddVariable_ScopeFailure() {
	r.withScope(errors.New("bacon"))

	ret, err := r.addVariable()

	r.Nil(ret)
	r.EqualError(err, "could not retrieve scope: bacon")
}

func (r *rootResolverTestSuite) TestAddVariable_AddFailure() {
	r.withScope(nil)

	r.withCreateVariable(
		"workspace/scopeName",
		&ssmvars.Variable{Name: "name", Value: "value", WriteOnly: true},
		errors.New("bacon"),
	)

	ret, err := r.addVariable()

	r.Nil(ret)
	r.EqualError(err, "could not create variable: bacon")
}

func (r *rootResolverTestSuite) TestRemoveVariable_OK() {
	variable := &ssmvars.Variable{}
	r.withDeleteVariable(variable, nil)

	ret, err := r.sut.RemoveVariable(r.ctx, removeVariableArgs{
		ScopeID: "scopeName",
		ID:      "variable",
	})

	r.NoError(err)
	r.Equal(variable, ret.wraps)
}

func (r *rootResolverTestSuite) TestRemoveVariable_BackendFailure() {
	r.withDeleteVariable((*ssmvars.Variable)(nil), errors.New("bacon"))

	ret, err := r.sut.RemoveVariable(r.ctx, removeVariableArgs{
		ScopeID: "scopeName",
		ID:      "variable",
	})

	r.Nil(ret)
	r.EqualError(err, "could not remove variable: bacon")
}

func (r *rootResolverTestSuite) TestCreateRelease_OK() {
	variable := &ssmvars.Variable{Name: "VARIABLE", Value: "value"}

	r.withScope(nil)
	r.withListVariables("workspace/scopeName", nil, variable)
	r.withCreateRelease(nil, variable)

	ret, err := r.sut.CreateRelease(r.ctx, scopeArgs{ScopeID: "scopeName"})

	r.NoError(err)
	r.EqualValues("releaseID", ret.ID())
	r.Equal(r.backend, ret.backend)
	r.NotNil(ret.scope)
}

func (r *rootResolverTestSuite) TestCreateRelease_ScopeError() {
	r.withScope(errors.New("bacon"))

	ret, err := r.sut.CreateRelease(r.ctx, scopeArgs{ScopeID: "scopeName"})

	r.Nil(ret)
	r.EqualError(err, "could not retrieve scope: bacon")
}

func (r *rootResolverTestSuite) TestCreateRelease_ListError() {
	r.withScope(nil)
	r.withListVariables("workspace/scopeName", errors.New("bacon"))

	ret, err := r.sut.CreateRelease(r.ctx, scopeArgs{ScopeID: "scopeName"})

	r.Nil(ret)
	r.EqualError(err, "could not list variables: bacon")
}

func (r *rootResolverTestSuite) TestCreateRelease_CreateError() {
	variable := &ssmvars.Variable{Name: "VARIABLE", Value: "value"}

	r.withScope(nil)
	r.withListVariables("workspace/scopeName", nil, variable)
	r.withCreateRelease(errors.New("bacon"), variable)

	ret, err := r.sut.CreateRelease(r.ctx, scopeArgs{ScopeID: "scopeName"})

	r.Nil(ret)
	r.EqualError(err, "could not create a release: bacon")
}

func (r *rootResolverTestSuite) TestArchiveRelease_OK() {
	r.withScope(nil)
	r.withArchiveRelease(nil)

	ret, err := r.sut.ArchiveRelease(r.ctx, mutateReleaseArgs{
		ScopeID:   "scopeName",
		ReleaseID: "releaseID",
	})

	r.NoError(err)
	r.EqualValues("releaseID", ret.ID())
	r.Equal(r.backend, ret.backend)
	r.NotNil(ret.scope)
}

func (r *rootResolverTestSuite) TestArchiveRelease_ScopeError() {
	r.withScope(errors.New("bacon"))

	ret, err := r.sut.ArchiveRelease(r.ctx, mutateReleaseArgs{
		ScopeID:   "scopeName",
		ReleaseID: "releaseID",
	})

	r.Nil(ret)
	r.EqualError(err, "could not retrieve scope: bacon")
}

func (r *rootResolverTestSuite) TestArchiveRelease_ArchiveError() {
	r.withScope(nil)
	r.withArchiveRelease(errors.New("bacon"))

	ret, err := r.sut.ArchiveRelease(r.ctx, mutateReleaseArgs{
		ScopeID:   "scopeName",
		ReleaseID: "releaseID",
	})

	r.Nil(ret)
	r.EqualError(err, "could not archive release: bacon")
}

func (r *rootResolverTestSuite) TestReset_OK() {
	variable := &ssmvars.Variable{Name: "VARIABLE", Value: "value"}

	r.withScope(nil)
	r.withGetRelease(variable, nil)
	r.backend.On("Reset", r.ctx, "workspace/scopeName").Return(nil)
	r.withCreateVariable("workspace/scopeName", variable, nil)

	ret, err := r.sut.Reset(r.ctx, mutateReleaseArgs{
		ScopeID:   "scopeName",
		ReleaseID: "releaseID",
	})

	r.NoError(err)
	r.Equal(r.backend, ret.backend)
	r.NotNil(ret.wraps)
}

func (r *rootResolverTestSuite) TestReset_ScopeError() {
	r.withScope(errors.New("bacon"))

	ret, err := r.sut.Reset(r.ctx, mutateReleaseArgs{
		ScopeID:   "scopeName",
		ReleaseID: "releaseID",
	})

	r.Nil(ret)
	r.EqualError(err, "could not retrieve scope: bacon")
}

func (r *rootResolverTestSuite) TestReset_GetReleaseError() {
	variable := &ssmvars.Variable{Name: "VARIABLE", Value: "value"}

	r.withScope(nil)
	r.withGetRelease(variable, errors.New("bacon"))

	ret, err := r.sut.Reset(r.ctx, mutateReleaseArgs{
		ScopeID:   "scopeName",
		ReleaseID: "releaseID",
	})

	r.Nil(ret)
	r.EqualError(err, "could not get release: bacon")
}

func (r *rootResolverTestSuite) TestReset_ResetError() {
	variable := &ssmvars.Variable{Name: "VARIABLE", Value: "value"}

	r.withScope(nil)
	r.withGetRelease(variable, nil)
	r.backend.On("Reset", r.ctx, "workspace/scopeName").Return(errors.New("bacon"))

	ret, err := r.sut.Reset(r.ctx, mutateReleaseArgs{
		ScopeID:   "scopeName",
		ReleaseID: "releaseID",
	})

	r.Nil(ret)
	r.EqualError(err, "could not reset the current workspace: bacon")
}

func (r *rootResolverTestSuite) TestReset_CreateVariableError() {
	variable := &ssmvars.Variable{Name: "VARIABLE", Value: "value"}

	r.withScope(nil)
	r.withGetRelease(variable, nil)
	r.backend.On("Reset", r.ctx, "workspace/scopeName").Return(nil)
	r.withCreateVariable("workspace/scopeName", variable, errors.New("bacon"))

	ret, err := r.sut.Reset(r.ctx, mutateReleaseArgs{
		ScopeID:   "scopeName",
		ReleaseID: "releaseID",
	})

	r.Nil(ret)
	r.EqualError(err, "could not create variable: bacon")
}

func (r *rootResolverTestSuite) addVariable() (*variableResolver, error) {
	return r.sut.AddVariable(r.ctx, addVariableArgs{
		ScopeID: "scopeName",
		Variable: variableInput{
			Name:      "name",
			Value:     "value",
			WriteOnly: true,
		},
	})
}

func (r *rootResolverTestSuite) withArchiveRelease(err error) {
	r.backend.On("ArchiveRelease", r.ctx, "scopeName", "releaseID").Return(err)
}

func (r *rootResolverTestSuite) withCreateRelease(err error, variables ...*ssmvars.Variable) {
	var ret *secretservice.Release

	if err == nil {
		ret = &secretservice.Release{
			ID:        "releaseID",
			ScopeName: "scopeName",
			Variables: variables,
		}
	}

	r.backend.On("CreateRelease", r.ctx, "scopeName", variables).Return(ret, err)
}

func (r *rootResolverTestSuite) withCreateVariable(namespace string, variable *ssmvars.Variable, err error) {
	r.backend.On(
		"CreateVariable",
		r.ctx,
		namespace,
		mock.MatchedBy(func(arg interface{}) bool {
			input, ok := arg.(*ssmvars.Variable)
			r.True(ok)

			r.Equal(variable.Name, input.Name)
			r.Equal(variable.Value, input.Value)
			r.Equal(variable.WriteOnly, input.WriteOnly)

			return true
		}),
	).Return(variable, err)
}

func (r *rootResolverTestSuite) withDeleteVariable(variable *ssmvars.Variable, err error) {
	r.backend.
		On("DeleteVariable", r.ctx, "workspace/scopeName", "variable").
		Return(variable, err)
}

func (r *rootResolverTestSuite) withGetRelease(variable *ssmvars.Variable, err error) {
	r.backend.On("GetRelease", r.ctx, "scopeName", "releaseID").Return(&secretservice.Release{
		ID:        "releaseID",
		ScopeName: "scopeName",
		Variables: []*ssmvars.Variable{variable},
	}, err)
}

func (r *rootResolverTestSuite) withListVariables(prefix string, err error, variables ...*ssmvars.Variable) {
	r.backend.On("ListVariables", r.ctx, prefix).Return(variables, err)
}

func (r *rootResolverTestSuite) withScope(err error) {
	ret := &secretservice.Scope{}

	if err == nil {
		ret.Name = "scopeName"
		ret.KMSKeyID = "kmsKeyID"
	}

	r.backend.On("Scope", r.ctx, "scopeName").Return(ret, err)
}

func TestRootResolver(t *testing.T) {
	suite.Run(t, new(rootResolverTestSuite))
}
