package dao

import (
	"time"
)

// Migration defines a migrations table record.
type Migration struct {
	Tag       string    `db:"tag"`
	Timestamp time.Time `db:"timestamp"`
}
