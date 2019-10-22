package types

import (
	"bytes"

	"github.com/trivigy/migrate/v2/global"
)

// Direction defines the type of the migration direction.
type Direction int

const (
	// DirectionUp indicates the direction type is forward.
	DirectionUp Direction = iota + 1

	// DirectionDown indicates the direction type is backward (rollback).
	DirectionDown
)

var (
	directionUpStr   = "up"
	directionDownStr = "down"
)

func (r Direction) String() string {
	return toStringDirection[r]
}

var toStringDirection = map[Direction]string{
	Direction(0):  global.UnknownStr,
	DirectionUp:   directionUpStr,
	DirectionDown: directionDownStr,
}

// MarshalJSON marshals the enum as a quoted json string.
func (r Direction) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	if _, err := buffer.WriteString(toStringDirection[r]); err != nil {
		return nil, err
	}
	if _, err := buffer.WriteString(`"`); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
