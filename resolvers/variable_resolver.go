package resolvers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/graph-gophers/graphql-go"
	"github.com/marcinwyszynski/ssmvars"
)

type variableResolver struct {
	wraps *ssmvars.Variable
}

// id: ID!
func (v *variableResolver) ID() graphql.ID {
	return graphql.ID(v.wraps.Name)
}

// value: String
func (v *variableResolver) Value() *string {
	if v.wraps.WriteOnly {
		return nil
	}
	return aws.String(v.wraps.Value)
}

// writeOnly: Boolean!
func (v *variableResolver) WriteOnly() bool {
	return v.wraps.WriteOnly
}
