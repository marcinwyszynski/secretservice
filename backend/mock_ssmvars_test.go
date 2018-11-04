package backend_test

import (
	"context"

	"github.com/marcinwyszynski/ssmvars"
	"github.com/stretchr/testify/mock"
)

type mockSSMVars struct {
	mock.Mock
	ssmvars.ReadWriter
}

func (m *mockSSMVars) ShowVariable(ctx context.Context, namespace, name string) (*ssmvars.Variable, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(*ssmvars.Variable), args.Error(1)
}
