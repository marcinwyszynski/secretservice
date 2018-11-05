package resolver

import (
	"context"
	"fmt"
	"path"

	"github.com/graph-gophers/graphql-go"
	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/ssmvars"
	"github.com/pkg/errors"
)

type rootResolver struct {
	wraps secretservice.Backend
}

// New returns an implementation of GraphQL resolver.
func New(backend secretservice.Backend) interface{} {
	return &rootResolver{wraps: backend}
}

type scopeArgs struct {
	ScopeID graphql.ID
}

// scope(scopeId: ID!): Scope!
func (r *rootResolver) Scope(ctx context.Context, args scopeArgs) (*scopeResolver, error) {
	scope, err := r.wraps.Scope(ctx, string(args.ScopeID))
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve scope")
	}

	return &scopeResolver{backend: r.wraps, wraps: scope}, nil
}

type createScopeArgs struct {
	Name, KMSKeyID string
}

// createScope(name: String!, kmsKeyId: String!): Scope!
func (r *rootResolver) CreateScope(ctx context.Context, args createScopeArgs) (*scopeResolver, error) {
	scopeName := args.Name
	keyID := args.KMSKeyID

	scopeKeys, err := r.wraps.ListVariables(ctx, "scopes")
	if err != nil {
		return nil, errors.Wrap(err, "could not list scopes")
	}
	for _, scope := range scopeKeys {
		if scope.Name == scopeName {
			return nil, errors.Errorf("scope %q already exists", scopeName)
		}
	}

	_, err = r.wraps.CreateVariable(ctx, "scopes", &ssmvars.Variable{Name: scopeName, Value: keyID})
	if err != nil {
		return nil, errors.Wrap(err, "could not create scope")
	}

	scope := &secretservice.Scope{Name: scopeName, KMSKeyID: keyID}
	return &scopeResolver{backend: r.wraps, wraps: scope}, nil
}

type variableInput struct {
	Name, Value string
	WriteOnly   bool
}

func (v variableInput) toSSM() *ssmvars.Variable {
	return &ssmvars.Variable{
		Name:      v.Name,
		Value:     v.Value,
		WriteOnly: v.WriteOnly,
	}
}

type addVariableArgs struct {
	ScopeID  graphql.ID
	Variable variableInput
}

// addVariable(scopeId: ID!, variable: VariableInput!): Variable!
func (r *rootResolver) AddVariable(ctx context.Context, args addVariableArgs) (*variableResolver, error) {
	scope, err := r.wraps.Scope(ctx, string(args.ScopeID))
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve scope")
	}

	variable, err := r.wraps.CreateVariable(
		ctx,
		path.Join("workspace", scope.Name),
		args.Variable.toSSM(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create variable")
	}

	return &variableResolver{wraps: variable}, nil
}

type removeVariableArgs struct {
	ScopeID graphql.ID
	ID      graphql.ID
}

// removeVariable(scopeId: ID!, id: ID!): Variable!
func (r *rootResolver) RemoveVariable(ctx context.Context, args removeVariableArgs) (*variableResolver, error) {
	variable, err := r.wraps.DeleteVariable(ctx, fmt.Sprintf("workspace/%s", args.ScopeID), string(args.ID))
	if err != nil {
		return nil, errors.Wrap(err, "could not remove variable")
	}

	return &variableResolver{wraps: variable}, nil
}

// createRelease(scopeId: ID!): Release!
func (r *rootResolver) CreateRelease(ctx context.Context, args scopeArgs) (*releaseResolver, error) {
	scopeName := string(args.ScopeID)

	scope, err := r.wraps.Scope(ctx, scopeName)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve scope")
	}

	variables, err := r.wraps.ListVariables(ctx, fmt.Sprintf("workspace/%s", scope.Name))
	if err != nil {
		return nil, errors.Wrap(err, "could not list variables")
	}

	release, err := r.wraps.CreateRelease(ctx, scopeName, variables)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a release")
	}

	return newReleaseResolver(r.wraps, graphql.ID(release.ID), scope), nil
}

type mutateReleaseArgs struct {
	ScopeID, ReleaseID graphql.ID
}

// archiveRelease(scopeId: ID!, releaseId: ID!): Release!
func (r *rootResolver) ArchiveRelease(ctx context.Context, args mutateReleaseArgs) (*releaseResolver, error) {
	scope, err := r.wraps.Scope(ctx, string(args.ScopeID))
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve scope")
	}

	if err := r.wraps.ArchiveRelease(ctx, scope.Name, string(args.ReleaseID)); err != nil {
		return nil, errors.Wrap(err, "could not archive release")
	}

	return newReleaseResolver(r.wraps, args.ReleaseID, scope), nil
}

// reset(scopeId: ID!, releaseId: ID!): Scope!
func (r *rootResolver) Reset(ctx context.Context, args mutateReleaseArgs) (*scopeResolver, error) {
	scopeName := string(args.ScopeID)

	scope, err := r.wraps.Scope(ctx, scopeName)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve scope")
	}

	release, err := r.wraps.GetRelease(ctx, scopeName, string(args.ReleaseID))
	if err != nil {
		return nil, errors.Wrap(err, "could not get release")
	}

	namespace := fmt.Sprintf("workspace/%s", scopeName)

	if err := r.wraps.Reset(ctx, namespace); err != nil {
		return nil, errors.Wrap(err, "could not reset the current workspace")
	}

	for _, variable := range release.Variables {
		if _, err := r.wraps.CreateVariable(ctx, namespace, variable); err != nil {
			return nil, errors.Wrap(err, "could not create variable")
		}
	}

	return &scopeResolver{backend: r.wraps, wraps: scope}, nil
}
