package resolver

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/ssmvars"
	"github.com/pkg/errors"
)

type resolver struct {
	wraps secretservice.Backend
}

// New returns an implementation of GraphQL resolver.
func New(backend secretservice.Backend) interface{} {
	return &resolver{wraps: backend}
}

type releaseArgs struct {
	ScopeID, ReleaseID graphql.ID
}

// release(scopeId: ID!, releaseId: ID!): Release!
func (r *resolver) Release(ctx context.Context, args releaseArgs) (*releaseResolver, error) {
	scopeName := string(args.ScopeID)

	scope, err := r.wraps.Scope(ctx, scopeName)
	if err != nil {
		return nil, err
	}

	release, err := r.wraps.GetRelease(ctx, scopeName, string(args.ReleaseID))
	if err != nil {
		return nil, errors.Wrap(err, "could not get release")
	}

	return &releaseResolver{scope: scope, wraps: release}, nil
}

type scopeArgs struct {
	ScopeID graphql.ID
}

// scope(scopeId: ID!): Scope!
func (r *resolver) Scope(ctx context.Context, args scopeArgs) (*scopeResolver, error) {
	scope, err := r.wraps.Scope(ctx, string(args.ScopeID))
	if err != nil {
		return nil, err
	}

	return &scopeResolver{backend: r.wraps, wraps: scope}, nil
}

type createScopeArgs struct {
	Name, KMSKeyID string
}

// createScope(name: String!, kmsKeyId: String!): Scope!
func (r *resolver) CreateScope(ctx context.Context, args createScopeArgs) (*scopeResolver, error) {
	scopeName := args.Name
	keyID := args.KMSKeyID

	scopes, err := r.wraps.ListVariables(ctx, "scopes")
	if err != nil {
		return nil, errors.Wrap(err, "could not list scopes")
	}
	for _, scope := range scopes {
		if scope.Value == scopeName {
			return nil, errors.Errorf("scope %s already exists", scopeName)
		}
	}

	_, err = r.wraps.CreateVariable(ctx, "scopes", &ssmvars.Variable{Name: scopeName, Value: keyID})
	if err != nil {
		return nil, errors.Wrap(err, "could not create scope")
	}

	return &scopeResolver{wraps: &secretservice.Scope{Name: scopeName, KMSKeyID: keyID}}, nil
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
func (r *resolver) AddVariable(ctx context.Context, args addVariableArgs) (*variableResolver, error) {
	variable, err := r.wraps.CreateVariable(
		ctx,
		fmt.Sprintf("workspace/%s", args.ScopeID),
		args.Variable.toSSM(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not get variable")
	}

	return &variableResolver{wraps: variable}, nil
}

type removeVariableArgs struct {
	ScopeID graphql.ID
	ID      graphql.ID
}

// removeVariable(scopeId: ID!, id: ID!): Variable!
func (r *resolver) RemoveVariable(ctx context.Context, args removeVariableArgs) (*variableResolver, error) {
	variable, err := r.wraps.DeleteVariable(ctx, fmt.Sprintf("workspace/%s", args.ScopeID), string(args.ID))
	if err != nil {
		return nil, errors.Wrap(err, "could not remove variable")
	}

	return &variableResolver{wraps: variable}, nil
}

// createRelease(scopeId: ID!): Release!
func (r *resolver) CreateRelease(ctx context.Context, args scopeArgs) (*releaseResolver, error) {
	scopeName := string(args.ScopeID)

	scope, err := r.wraps.Scope(ctx, scopeName)
	if err != nil {
		return nil, err
	}

	variables, err := r.wraps.ListVariables(ctx, fmt.Sprintf("workspace/%s", scope.Name))
	if err != nil {
		return nil, errors.Wrap(err, "could not list variables")
	}

	release, err := r.wraps.CreateRelease(ctx, scopeName, variables)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a release")
	}

	return &releaseResolver{wraps: release}, nil
}

// showRelease(scopeId: ID!, releaseId: ID!): Release!
func (r *resolver) ShowRelease(ctx context.Context, args releaseArgs) (*releaseResolver, error) {
	release, err := r.wraps.GetRelease(ctx, string(args.ScopeID), string(args.ReleaseID))
	if err != nil {
		return nil, errors.Wrap(err, "could not get release")
	}

	return &releaseResolver{wraps: release}, nil
}

// archiveRelease(scopeId: ID!, releaseId: ID!): Release!
func (r *resolver) ArchiveRelease(ctx context.Context, args releaseArgs) (*releaseResolver, error) {
	ret, err := r.ShowRelease(ctx, args)
	if err != nil {
		return nil, err
	}

	if err := r.wraps.ArchiveRelease(ctx, string(args.ScopeID), string(args.ReleaseID)); err != nil {
		return nil, errors.Wrap(err, "could not archive release")
	}

	return ret, nil
}

// reset(scopeId: ID!, releaseId: ID!): Scope!
func (r *resolver) Reset(ctx context.Context, args releaseArgs) (*scopeResolver, error) {
	scopeName := string(args.ScopeID)

	scope, err := r.wraps.Scope(ctx, scopeName)
	if err != nil {
		return nil, err
	}

	release, err := r.wraps.GetRelease(ctx, scopeName, string(args.ReleaseID))
	if err != nil {
		return nil, errors.Wrap(err, "could not get release")
	}

	namespace := fmt.Sprintf("workspace/%s", scopeName)

	if err := r.wraps.Reset(ctx, namespace); err != nil {
		return nil, errors.Wrap(err, "could not clean the current workspace")
	}

	for _, variable := range release.Variables {
		if _, err := r.wraps.CreateVariable(ctx, namespace, variable); err != nil {
			return nil, errors.Wrap(err, "could not create variable")
		}
	}

	return &scopeResolver{backend: r.wraps, wraps: scope}, nil
}
