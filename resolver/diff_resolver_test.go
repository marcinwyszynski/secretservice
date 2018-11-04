package resolver

import (
	"testing"

	"github.com/marcinwyszynski/ssmvars"
	"github.com/stretchr/testify/suite"
)

var (
	unchangedVariable = &ssmvars.Variable{Name: "UNCHANGED"}

	addedVariable = &ssmvars.Variable{Name: "NEW"}

	deletedVariable = &ssmvars.Variable{Name: "OLD"}

	changedVariableWriteOld = &ssmvars.Variable{Name: "CHANGED_WRITE", WriteOnly: true}
	changedVariableWriteNew = &ssmvars.Variable{Name: "CHANGED_WRITE", WriteOnly: false}

	changedVariableValueOld = &ssmvars.Variable{Name: "CHANGED_VALUE", Value: "old"}
	changedVariableValueNew = &ssmvars.Variable{Name: "CHANGED_VALUE", Value: "new"}
)

type diffResolverTestSuite struct {
	suite.Suite

	sut *diffResolver
}

func (d *diffResolverTestSuite) SetupTest() {
	oldVariables := []*ssmvars.Variable{
		unchangedVariable,
		deletedVariable,
		changedVariableWriteOld,
		changedVariableValueOld,
	}

	newVariables := []*ssmvars.Variable{
		unchangedVariable,
		addedVariable,
		changedVariableWriteNew,
		changedVariableValueNew,
	}

	d.sut = newDiffResolver(oldVariables, newVariables)
}

func (d *diffResolverTestSuite) TestAdded() {
	added := d.sut.Added()

	d.Len(added, 1)
	d.Equal(addedVariable, added[0].wraps)
}

func (d *diffResolverTestSuite) TestChanged() {
	changed := d.sut.Changed()

	d.Len(changed, 2)
}

func (d *diffResolverTestSuite) TestDeleted() {
	deleted := d.sut.Deleted()

	d.Len(deleted, 1)
	d.Equal(deletedVariable, deleted[0].wraps)
}

func TestDiffResolver(t *testing.T) {
	suite.Run(t, new(diffResolverTestSuite))
}
