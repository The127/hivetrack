package handlers

import (
	"encoding/json"

	"github.com/google/uuid"
)

// parseNullableInt decodes a json.RawMessage as a nullable int using **int
// patch semantics:
//   - nil/empty raw → nil (field absent, no update)
//   - "null"        → &(*int=nil) (explicitly clear)
//   - number        → &(*int=&v) (set to value)
func parseNullableInt(raw json.RawMessage) (**int, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if string(raw) == "null" {
		p := (*int)(nil)
		return &p, nil
	}
	var v int
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	ptr := &v
	return &ptr, nil
}

// parseNullableUUID decodes a json.RawMessage as a nullable UUID.
// Returns (value, clear, err) where:
//   - absent field → (nil, false, nil) — no update
//   - "null"       → (nil, true, nil)  — explicitly clear
//   - valid UUID   → (&id, false, nil) — set to value
func parseNullableUUID(raw json.RawMessage) (value *uuid.UUID, clear bool, err error) {
	if len(raw) == 0 {
		return nil, false, nil
	}
	if string(raw) == "null" {
		return nil, true, nil
	}
	var id uuid.UUID
	if err := json.Unmarshal(raw, &id); err != nil {
		return nil, false, err
	}
	return &id, false, nil
}
