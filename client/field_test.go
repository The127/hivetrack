package client

import (
	"encoding/json"
	"testing"
)

// --- Field[T] unit tests ---

func TestField_zeroValue_isAbsent(t *testing.T) {
	var f Field[string]
	if !f.IsAbsent() {
		t.Error("zero value should be absent")
	}
	if f.IsSet() || f.IsNull() {
		t.Error("zero value should not be set or null")
	}
}

func TestField_Set_isSet(t *testing.T) {
	f := Set("hello")
	if !f.IsSet() {
		t.Error("Set field should be set")
	}
	if f.IsAbsent() || f.IsNull() {
		t.Error("Set field should not be absent or null")
	}
	if f.Value() != "hello" {
		t.Errorf("expected hello, got %s", f.Value())
	}
}

func TestField_Null_isNull(t *testing.T) {
	f := Null[string]()
	if !f.IsNull() {
		t.Error("Null field should be null")
	}
	if f.IsAbsent() || f.IsSet() {
		t.Error("Null field should not be absent or set")
	}
}

func TestField_MarshalJSON_set(t *testing.T) {
	f := Set("hello")
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"hello"` {
		t.Errorf("expected \"hello\", got %s", data)
	}
}

func TestField_MarshalJSON_null(t *testing.T) {
	f := Null[string]()
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Errorf("expected null, got %s", data)
	}
}

func TestField_MarshalJSON_int(t *testing.T) {
	f := Set(42)
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "42" {
		t.Errorf("expected 42, got %s", data)
	}
}

func TestField_MarshalJSON_bool(t *testing.T) {
	f := Set(true)
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "true" {
		t.Errorf("expected true, got %s", data)
	}
}

func TestField_MarshalJSON_slice(t *testing.T) {
	f := Set([]string{"a", "b"})
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `["a","b"]` {
		t.Errorf("expected [\"a\",\"b\"], got %s", data)
	}
}

func TestField_MarshalJSON_nullSlice(t *testing.T) {
	f := Null[[]string]()
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Errorf("expected null, got %s", data)
	}
}

// --- marshalFields edge cases ---

func TestMarshalFields_emptyStruct(t *testing.T) {
	type empty struct{}
	data, err := marshalFields(empty{})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}" {
		t.Errorf("expected {}, got %s", data)
	}
}

func TestMarshalFields_allAbsent(t *testing.T) {
	req := UpdateIssueRequest{} // all fields absent
	data, err := marshalFields(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}" {
		t.Errorf("expected {}, got %s", data)
	}
}

func TestMarshalFields_mixedRegularAndFieldTypes(t *testing.T) {
	// BatchUpdateIssuesRequest has both regular (Numbers []int) and Field[T] fields
	req := BatchUpdateIssuesRequest{
		Numbers:  []int{1, 2},
		Status:   Set("done"),
		Priority: Null[string](),
		// Estimate is absent
	}
	data, err := marshalFields(req)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)

	// Regular field included
	nums := m["numbers"].([]any)
	if len(nums) != 2 {
		t.Errorf("expected 2 numbers, got %d", len(nums))
	}
	// Set field included
	if m["status"] != "done" {
		t.Errorf("expected status=done, got %v", m["status"])
	}
	// Null field included as null
	v, ok := m["priority"]
	if !ok {
		t.Fatal("expected priority in map")
	}
	if v != nil {
		t.Errorf("expected null priority, got %v", v)
	}
	// Absent field omitted
	if _, ok := m["estimate"]; ok {
		t.Error("absent estimate should not be in JSON")
	}
}

func TestMarshalFields_regularZeroFieldOmitted(t *testing.T) {
	req := BatchUpdateIssuesRequest{
		Status: Set("todo"),
		// Numbers is nil (zero value for []int) — should be omitted
	}
	data, err := marshalFields(req)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)
	if _, ok := m["numbers"]; ok {
		t.Error("nil numbers should be omitted")
	}
}

func TestMarshalFields_fieldWithNoJsonTag_skipped(t *testing.T) {
	type hasUntagged struct {
		Name    Field[string] `json:"name"`
		Private Field[string] // no json tag
		Ignored Field[string] `json:"-"`
	}
	req := hasUntagged{
		Name:    Set("test"),
		Private: Set("secret"),
		Ignored: Set("skip"),
	}
	data, err := marshalFields(req)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["name"] != "test" {
		t.Errorf("expected name=test, got %v", m["name"])
	}
	if len(m) != 1 {
		t.Errorf("expected only 1 key, got %d: %v", len(m), m)
	}
}

func TestMarshalFields_pointer(t *testing.T) {
	req := UpdateIssueRequest{Title: Set("via pointer")}
	// marshalFields should handle *struct as well as struct
	data, err := marshalFields(&req)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["title"] != "via pointer" {
		t.Errorf("expected title, got %v", m["title"])
	}
}
