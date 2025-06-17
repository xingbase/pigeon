package id

import (
	"strconv"
	"time"

	"github.com/xingbase/pigeon"
)

// tm generates an id based on current time
type tm struct {
	Now func() time.Time
}

// NewTime builds a pigeon.ID generator based on current time
func NewTime() pigeon.ID {
	return &tm{
		Now: time.Now,
	}
}

// Generate creates a string based on the current time as an integer
func (i *tm) Generate() string {
	return strconv.Itoa(int(i.Now().Unix()))
}
