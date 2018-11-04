package resolver

import (
	"testing"

	"github.com/marcinwyszynski/ssmvars"
	"github.com/stretchr/testify/suite"
)

type variableResolverTestSuite struct {
	suite.Suite

	variable *ssmvars.Variable
	sut      *variableResolver
}

func (v *variableResolverTestSuite) SetupTest() {
	v.variable = &ssmvars.Variable{Name: "NAME", Value: "value"}
	v.sut = &variableResolver{wraps: v.variable}
}

func (v *variableResolverTestSuite) TestID() {
	v.EqualValues("NAME", v.sut.ID())
}

func (v *variableResolverTestSuite) TestValue_Public() {
	v.variable.WriteOnly = false
	v.Equal("value", *v.sut.Value())
}

func (v *variableResolverTestSuite) TestValue_Secret() {
	v.variable.WriteOnly = true
	v.Nil(v.sut.Value())
}

func (v *variableResolverTestSuite) TestWriteOnly_Public() {
	v.variable.WriteOnly = false
	v.False(v.sut.WriteOnly())
}

func (v *variableResolverTestSuite) TestWriteOnly_Secret() {
	v.variable.WriteOnly = true
	v.True(v.sut.WriteOnly())
}

func TestVariableResolver(t *testing.T) {
	suite.Run(t, new(variableResolverTestSuite))
}
