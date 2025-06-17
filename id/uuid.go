package id

import (
	"github.com/google/uuid"
	"github.com/xingbase/pigeon"
)

var _ pigeon.ID = &UUID{}

// UUID generates a V4 uuid
type UUID struct{}

// Generate creates a UUID v4 string
func (i *UUID) Generate() string {
	return uuid.New().String()
}
