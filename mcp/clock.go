package mcp

import "time"

// Clock abstracts time.Now() to allow deterministic tests.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// RealClock is the production Clock implementation.
var RealClock Clock = realClock{}
