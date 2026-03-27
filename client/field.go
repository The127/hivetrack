package client

import (
	"encoding/json"
	"reflect"
)

// Field represents an optional API field with three states:
//   - Absent: not included in JSON (zero value of Field)
//   - Null: explicitly set to null in JSON (clears the field server-side)
//   - Set: carries a value
type Field[T any] struct {
	value T
	set   bool
	null  bool
}

// Set creates a Field with a value.
func Set[T any](v T) Field[T] {
	return Field[T]{value: v, set: true}
}

// Null creates a Field that marshals as JSON null.
func Null[T any]() Field[T] {
	return Field[T]{null: true}
}

// IsSet reports whether the field carries a value.
func (f Field[T]) IsSet() bool { return f.set }

// IsNull reports whether the field should be serialized as null.
func (f Field[T]) IsNull() bool { return f.null }

// IsAbsent reports whether the field should be omitted from JSON.
func (f Field[T]) IsAbsent() bool { return !f.set && !f.null }

// Value returns the field's value. Only meaningful when IsSet() is true.
func (f Field[T]) Value() T { return f.value }

// MarshalJSON implements json.Marshaler.
func (f Field[T]) MarshalJSON() ([]byte, error) {
	if f.null {
		return []byte("null"), nil
	}
	return json.Marshal(f.value)
}

// marshalFields marshals a struct containing Field[T] values into JSON,
// omitting absent fields. Struct fields must have a `json` tag.
//
// This replaces the toMap() pattern — instead of building a map manually,
// each request struct is marshaled directly using reflection that understands
// the Field[T] three-state semantics.
func marshalFields(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()

	m := make(map[string]any)

	for i := range rt.NumField() {
		field := rt.Field(i)
		fv := rv.Field(i)

		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		// Strip ",omitempty" or other options from tag.
		name := tag
		if idx := indexOf(tag, ','); idx != -1 {
			name = tag[:idx]
		}

		// Check if this is a Field[T] (has IsAbsent/IsNull/IsSet methods).
		if isFieldType(fv) {
			if callBool(fv, "IsAbsent") {
				continue
			}
			if callBool(fv, "IsNull") {
				m[name] = nil
				continue
			}
			// Get the underlying value via MarshalJSON.
			marshaler := fv.MethodByName("MarshalJSON")
			if marshaler.IsValid() {
				results := marshaler.Call(nil)
				var raw json.RawMessage
				raw = results[0].Bytes()
				if results[1].Interface() != nil {
					return nil, results[1].Interface().(error)
				}
				m[name] = raw
				continue
			}
		}

		// Regular fields: include if non-zero (like omitempty).
		if fv.IsZero() {
			continue
		}
		m[name] = fv.Interface()
	}

	return json.Marshal(m)
}

func isFieldType(v reflect.Value) bool {
	return v.MethodByName("IsAbsent").IsValid() &&
		v.MethodByName("IsNull").IsValid() &&
		v.MethodByName("IsSet").IsValid()
}

func callBool(v reflect.Value, method string) bool {
	m := v.MethodByName(method)
	if !m.IsValid() {
		return false
	}
	results := m.Call(nil)
	if len(results) == 1 && results[0].Kind() == reflect.Bool {
		return results[0].Bool()
	}
	return false
}

func indexOf(s string, c byte) int {
	for i := range len(s) {
		if s[i] == c {
			return i
		}
	}
	return -1
}
