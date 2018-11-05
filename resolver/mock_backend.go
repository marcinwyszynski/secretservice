package resolver

import (
	"context"

	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/ssmvars"
	"github.com/stretchr/testify/mock"
)

type mockBackend struct {
	mock.Mock
	secretservice.Backend
}

func (m *mockBackend) ArchiveRelease(ctx context.Context, scopeName, releaseID string) error {
	return m.Called(ctx, scopeName, releaseID).Error(0)
}

func (m *mockBackend) CreateRelease(ctx context.Context, scopeName string, variables []*ssmvars.Variable) (*secretservice.Release, error) {
	args := m.Called(ctx, scopeName, variables)
	return args.Get(0).(*secretservice.Release), args.Error(1)
}

func (m *mockBackend) GetRelease(ctx context.Context, scopeName, releaseID string) (*secretservice.Release, error) {
	args := m.Called(ctx, scopeName, releaseID)
	return args.Get(0).(*secretservice.Release), args.Error(1)
}

func (m *mockBackend) ListReleases(ctx context.Context, scopeName string, before *string) ([]string, error) {
	args := m.Called(ctx, scopeName, before)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockBackend) Scope(ctx context.Context, scopeName string) (*secretservice.Scope, error) {
	args := m.Called(ctx, scopeName)
	return args.Get(0).(*secretservice.Scope), args.Error(1)
}

func (m *mockBackend) ListVariables(ctx context.Context, namespace string) ([]*ssmvars.Variable, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).([]*ssmvars.Variable), args.Error(1)
}

func (m *mockBackend) Reset(ctx context.Context, namespace string) error {
	return m.Called(ctx, namespace).Error(0)
}

func (m *mockBackend) CreateVariable(ctx context.Context, namespace string, variable *ssmvars.Variable) (*ssmvars.Variable, error) {
	args := m.Called(ctx, namespace, variable)
	return args.Get(0).(*ssmvars.Variable), args.Error(1)
}

func (m *mockBackend) DeleteVariable(ctx context.Context, namespace, name string) (*ssmvars.Variable, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(*ssmvars.Variable), args.Error(1)
}
