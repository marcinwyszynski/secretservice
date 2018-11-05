package resolver

import (
	"testing"

	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/marcinwyszynski/ssmvars"
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
