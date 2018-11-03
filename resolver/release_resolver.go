package resolver

import (
	"context"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marcinwyszynski/secretservice"
	"github.com/pkg/errors"
)

type releaseResolver struct {
	backend secretservice.Backend
	id      graphql.ID
	mutex   *sync.Mutex
	scope   *secretservice.Scope
	wraps   *secretservice.Release
}

func newReleaseResolver(backend secretservice.Backend, id graphql.ID, scope *secretservice.Scope) *releaseResolver {
	return &releaseResolver{
		backend: backend,
		id:      id,
		mutex:   new(sync.Mutex),
		scope:   scope,
	}
}

// id: ID!
func (r *releaseResolver) ID() graphql.ID {
	return graphql.ID(r.id)
}

// scope: Scope!
func (r *releaseResolver) Scope() *scopeResolver {
	return &scopeResolver{wraps: r.scope}
}

// live: Boolean!
func (r *releaseResolver) Live(ctx context.Context) (bool, error) {
	if err := r.loadRelease(ctx); err != nil {
		return false, err
	}

	return r.wraps.Live, nil
}

// variables: [Variable!]!
func (r *releaseResolver) Variables(ctx context.Context) ([]*variableResolver, error) {
	if err := r.loadRelease(ctx); err != nil {
		return nil, err
	}

	num := len(r.wraps.Variables)
	ret := make([]*variableResolver, num, num)
	for i, variable := range r.wraps.Variables {
		ret[i] = &variableResolver{wraps: variable}
	}
	return ret, nil
}

func (r *releaseResolver) loadRelease(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.wraps != nil {
		return nil
	}

	release, err := r.backend.GetRelease(ctx, r.scope.Name, string(r.id))
	if err != nil {
		return errors.Wrap(err, "could not lazily retrieve release")
	}
	r.wraps = release
	return nil
}
