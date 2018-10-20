package resolvers

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"

	"github.com/marcinwyszynski/secretservice"
)

type scopeResolver struct {
	backend secretservice.Backend
	wraps   *secretservice.Scope
}

// id: ID!
func (s *scopeResolver) ID() graphql.ID {
	return graphql.ID(s.wraps.Name)
}

// kmsKeyId: String!
func (s *scopeResolver) KMSKeyID() string {
	return s.wraps.KMSKeyID
}

// variables: [Variable!]!
func (s *scopeResolver) Variables(ctx context.Context) ([]*variableResolver, error) {
	variables, err := s.backend.ListVariables(ctx, fmt.Sprintf("workspace/%s", s.wraps.Name))
	if err != nil {
		return nil, errors.Wrap(err, "could not get workspace")
	}

	num := len(variables)
	ret := make([]*variableResolver, num, num)
	for i, variable := range variables {
		ret[i] = &variableResolver{wraps: variable}
	}
	return ret, nil
}
