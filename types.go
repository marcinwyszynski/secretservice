package secretservice

import "github.com/marcinwyszynski/ssmvars"

type Release struct {
	ID        string              `json:"-"`
	ScopeName string              `json:"-"`
	Live      bool                `json:"-"`
	Variables []*ssmvars.Variable `json:"variables"`
}

type Scope struct {
	Name     string `json:"-"`
	KMSKeyID string `json:"-"`
}
