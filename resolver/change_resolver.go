package resolver

import (
	"github.com/marcinwyszynski/ssmvars"
)

type changeResolver struct {
	before *ssmvars.Variable
	after  *ssmvars.Variable
}

// before: Variable!
func (c *changeResolver) Before() *variableResolver {
	return &variableResolver{wraps: c.before}
}

// after: Variable!
func (c *changeResolver) After() *variableResolver {
	return &variableResolver{wraps: c.after}
}
