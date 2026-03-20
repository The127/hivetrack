package mcp

import "time"

type fakeClock struct{ now time.Time }

func (f *fakeClock) Now() time.Time          { return f.now }
func (f *fakeClock) Tick(d time.Duration)    { f.now = f.now.Add(d) }
