package resolver

import (
	"testing"

	"github.com/marcinwyszynski/ssmvars"
	"github.com/stretchr/testify/assert"
)

func TestChangeResolver(t *testing.T) {
	before := &ssmvars.Variable{Name: "VAR", Value: "before"}
	after := &ssmvars.Variable{Name: "VAR", Value: "after"}

	resolver := &changeResolver{before: before, after: after}

	beforeResolver := resolver.Before()
	assert.Equal(t, before, beforeResolver.wraps)

	afterResolver := resolver.After()
	assert.Equal(t, after, afterResolver.wraps)
}
