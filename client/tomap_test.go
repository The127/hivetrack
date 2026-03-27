package client

import (
	"encoding/json"
	"testing"
)

func mustMarshal(t *testing.T, v any) map[string]any {
	t.Helper()
	data, err := marshalFields(v)
	if err != nil {
		t.Fatalf("marshalFields failed: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	return m
}

func TestUpdateIssueRequest_setsFields(t *testing.T) {
	req := UpdateIssueRequest{
		Title:    Set("New Title"),
		Status:   Set("done"),
		Priority: Set("high"),
		Estimate: Set("m"),
	}
	m := mustMarshal(t, req)
	if m["title"] != "New Title" {
		t.Errorf("expected title, got %v", m["title"])
	}
	if m["status"] != "done" {
		t.Errorf("expected status, got %v", m["status"])
	}
}

func TestUpdateIssueRequest_omitsAbsentFields(t *testing.T) {
	req := UpdateIssueRequest{Title: Set("Only Title")}
	m := mustMarshal(t, req)
	if _, ok := m["status"]; ok {
		t.Error("absent status should not be in JSON")
	}
	if _, ok := m["sprint_id"]; ok {
		t.Error("absent sprint_id should not be in JSON")
	}
}

func TestUpdateIssueRequest_nullClearsField(t *testing.T) {
	req := UpdateIssueRequest{SprintID: Null[string]()}
	m := mustMarshal(t, req)
	v, ok := m["sprint_id"]
	if !ok {
		t.Fatal("expected sprint_id in JSON")
	}
	if v != nil {
		t.Errorf("expected null, got %v", v)
	}
}

func TestUpdateIssueRequest_nullParentID(t *testing.T) {
	req := UpdateIssueRequest{ParentID: Null[string]()}
	m := mustMarshal(t, req)
	if m["parent_id"] != nil {
		t.Errorf("expected null, got %v", m["parent_id"])
	}
}

func TestUpdateIssueRequest_nullAssignees(t *testing.T) {
	req := UpdateIssueRequest{AssigneeIDs: Null[[]string]()}
	m := mustMarshal(t, req)
	if m["assignee_ids"] != nil {
		t.Errorf("expected null, got %v", m["assignee_ids"])
	}
}

func TestUpdateIssueRequest_nullLabels(t *testing.T) {
	req := UpdateIssueRequest{LabelIDs: Null[[]string]()}
	m := mustMarshal(t, req)
	if m["label_ids"] != nil {
		t.Errorf("expected null, got %v", m["label_ids"])
	}
}

func TestUpdateIssueRequest_nullOwnerID(t *testing.T) {
	req := UpdateIssueRequest{OwnerID: Null[string]()}
	m := mustMarshal(t, req)
	if m["owner_id"] != nil {
		t.Errorf("expected null, got %v", m["owner_id"])
	}
}

func TestUpdateIssueRequest_setAssigneeIDs(t *testing.T) {
	req := UpdateIssueRequest{AssigneeIDs: Set([]string{"a", "b"})}
	m := mustMarshal(t, req)
	raw, ok := m["assignee_ids"]
	if !ok {
		t.Fatal("expected assignee_ids")
	}
	// marshalFields produces json.RawMessage, which unmarshal reads as []any
	data, _ := json.Marshal(raw)
	var ids []string
	json.Unmarshal(data, &ids)
	if len(ids) != 2 || ids[0] != "a" {
		t.Errorf("unexpected: %v", ids)
	}
}

func TestUpdateSprintRequest_allFields(t *testing.T) {
	req := UpdateSprintRequest{
		Name:                     Set("Sprint 1"),
		Goal:                     Set("Ship it"),
		Status:                   Set("completed"),
		Force:                    Set(true),
		MoveOpenIssuesToSprintID: Set("next-uuid"),
	}
	m := mustMarshal(t, req)
	if m["name"] != "Sprint 1" {
		t.Errorf("expected name, got %v", m["name"])
	}
	if m["force"] != true {
		t.Errorf("expected force=true, got %v", m["force"])
	}
	if m["move_open_issues_to_sprint_id"] != "next-uuid" {
		t.Errorf("expected move id, got %v", m["move_open_issues_to_sprint_id"])
	}
}

func TestUpdateSprintRequest_omitsAbsent(t *testing.T) {
	req := UpdateSprintRequest{Name: Set("Only Name")}
	m := mustMarshal(t, req)
	if _, ok := m["force"]; ok {
		t.Error("absent force should not be in JSON")
	}
	if _, ok := m["status"]; ok {
		t.Error("absent status should not be in JSON")
	}
}

func TestUpdateProjectRequest_allFields(t *testing.T) {
	req := UpdateProjectRequest{
		Name:               Set("New Name"),
		Description:        Set("desc"),
		Archived:           Set(true),
		WipLimitInProgress: Set(5),
		WipLimitInReview:   Set(3),
	}
	m := mustMarshal(t, req)
	if m["name"] != "New Name" {
		t.Errorf("expected name, got %v", m["name"])
	}
	if m["archived"] != true {
		t.Errorf("expected archived=true, got %v", m["archived"])
	}
	// JSON numbers come back as float64
	if m["wip_limit_in_progress"] != float64(5) {
		t.Errorf("expected 5, got %v", m["wip_limit_in_progress"])
	}
}

func TestUpdateProjectRequest_nullWipLimit(t *testing.T) {
	req := UpdateProjectRequest{WipLimitInProgress: Null[int]()}
	m := mustMarshal(t, req)
	v, ok := m["wip_limit_in_progress"]
	if !ok {
		t.Fatal("expected wip_limit_in_progress in JSON")
	}
	if v != nil {
		t.Errorf("expected null, got %v", v)
	}
}

func TestUpdateMilestoneRequest_allFields(t *testing.T) {
	req := UpdateMilestoneRequest{
		Title: Set("v1.0"),
		Close: Set(true),
	}
	m := mustMarshal(t, req)
	if m["title"] != "v1.0" {
		t.Errorf("expected title, got %v", m["title"])
	}
	if m["close"] != true {
		t.Errorf("expected close=true, got %v", m["close"])
	}
}

func TestUpdateMilestoneRequest_reopenMilestone(t *testing.T) {
	req := UpdateMilestoneRequest{Close: Set(false)}
	m := mustMarshal(t, req)
	if m["close"] != false {
		t.Errorf("expected close=false, got %v", m["close"])
	}
}

func TestBatchUpdateIssuesRequest_includesNumbers(t *testing.T) {
	req := BatchUpdateIssuesRequest{
		Numbers: []int{1, 2, 3},
		Status:  Set("in_progress"),
	}
	m := mustMarshal(t, req)
	// numbers is a regular field, not Field[T]
	nums := m["numbers"].([]any)
	if len(nums) != 3 {
		t.Errorf("expected 3 numbers, got %d", len(nums))
	}
	if m["status"] != "in_progress" {
		t.Errorf("expected status, got %v", m["status"])
	}
}

func TestBatchUpdateIssuesRequest_clearSprintID(t *testing.T) {
	req := BatchUpdateIssuesRequest{
		Numbers:       []int{1},
		ClearSprintID: Set(true),
	}
	m := mustMarshal(t, req)
	if m["clear_sprint_id"] != true {
		t.Errorf("expected clear_sprint_id=true, got %v", m["clear_sprint_id"])
	}
}
