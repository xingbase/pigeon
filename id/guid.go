package id

import (
	"github.com/rs/xid"
	"github.com/xingbase/pigeon"
)

var _ pigeon.ID = &GUID{}

type GUID struct{}

// Generate creates a unique ID string
func (i *GUID) Generate() string {
	return xid.New().String()
}
