package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marcinwyszynski/secretservice"
)

type releaseResolver struct {
	scope *secretservice.Scope
	wraps *secretservice.Release
}

// id: ID!
func (r *releaseResolver) ID() graphql.ID {
	return graphql.ID(r.wraps.ID)
}

// scope: Scope!
func (r *releaseResolver) Scope() *scopeResolver {
	return &scopeResolver{wraps: r.scope}
}

// live: Boolean!
func (r *releaseResolver) Live() bool {
	return r.wraps.Live
}

// variables: [Variable!]!
func (r *releaseResolver) Variables() []*variableResolver {
	num := len(r.wraps.Variables)
	ret := make([]*variableResolver, num, num)
	for i, variable := range r.wraps.Variables {
		ret[i] = &variableResolver{wraps: variable}
	}
	return ret
}
