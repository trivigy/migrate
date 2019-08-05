package model

import (
	"time"
)

// Migration defines a migrations table record.
type Migration struct {
	Tag       string    `db:"tag"`
	Name      string    `db:"name"`
	Timestamp time.Time `db:"timestamp"`
}
