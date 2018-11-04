package resolver

import (
	"github.com/marcinwyszynski/ssmvars"
)

type diffResolver struct {
	oldVariables map[string]*ssmvars.Variable
	newVariables map[string]*ssmvars.Variable
}

func newDiffResolver(oldVariables, newVariables []*ssmvars.Variable) *diffResolver {
	ret := &diffResolver{
		oldVariables: make(map[string]*ssmvars.Variable),
		newVariables: make(map[string]*ssmvars.Variable),
	}

	for _, variable := range oldVariables {
		ret.oldVariables[variable.Name] = variable
	}
	for _, variable := range newVariables {
		ret.newVariables[variable.Name] = variable
	}

	return ret
}

// added: [Variable!]!
func (d *diffResolver) Added() []*variableResolver {
	return setDifference(d.newVariables, d.oldVariables)
}

// changed: [Change!]!
func (d *diffResolver) Changed() []*changeResolver {
	var ret []*changeResolver

	for key, before := range d.oldVariables {
		after, exists := d.newVariables[key]
		if !exists {
			continue
		}
		if before.Value != after.Value || before.WriteOnly != after.WriteOnly {
			ret = append(ret, &changeResolver{
				before: before,
				after:  after,
			})
		}
	}

	return ret
}

// deleted: [Variable!]!
func (d *diffResolver) Deleted() []*variableResolver {
	return setDifference(d.oldVariables, d.newVariables)
}

// setDifference returns the list of variables (wrapped in a resolver) which are
// in the first set, but not in the second one.
func setDifference(first, second map[string]*ssmvars.Variable) []*variableResolver {
	var ret []*variableResolver

	for key, variable := range first {
		if _, exists := second[key]; !exists {
			ret = append(ret, &variableResolver{wraps: variable})
		}

	}
	return ret
}
