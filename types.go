package secretservice

import (
	"github.com/marcinwyszynski/ssmvars"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
)

type Release struct {
	ID        string              `json:"-"`
	ScopeName string              `json:"-"`
	Live      bool                `json:"-"`
	Variables []*ssmvars.Variable `json:"variables"`
}

func (r *Release) Timestamp() (int64, error) {
	id, err := ulid.Parse(r.ID)
	if err != nil {
		return -1, errors.Wrap(err, "could not parse release ID as ULID")
	}

	millis := ulid.MaxTime() - id.Time()

	return int64(millis / 1e3), nil
}

type Scope struct {
	Name     string `json:"-"`
	KMSKeyID string `json:"-"`
}
