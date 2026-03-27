package client

import (
	"testing"
)

func ptr(s string) *string { return &s }
func boolPtr(b bool) *bool { return &b }
func intPtr(n int) *int    { return &n }

func TestUpdateIssueRequest_toMap_setsFields(t *testing.T) {
	req := UpdateIssueRequest{
		Title:    ptr("New Title"),
		Status:   ptr("done"),
		Priority: ptr("high"),
		Estimate: ptr("m"),
	}
	m := req.toMap()
	if m["title"] != "New Title" {
		t.Errorf("expected title, got %v", m["title"])
	}
	if m["status"] != "done" {
		t.Errorf("expected status, got %v", m["status"])
	}
	if m["priority"] != "high" {
		t.Errorf("expected priority, got %v", m["priority"])
	}
	if m["estimate"] != "m" {
		t.Errorf("expected estimate, got %v", m["estimate"])
	}
}

func TestUpdateIssueRequest_toMap_omitsNilFields(t *testing.T) {
	req := UpdateIssueRequest{Title: ptr("Only Title")}
	m := req.toMap()
	if _, ok := m["status"]; ok {
		t.Error("nil status should not be in map")
	}
	if _, ok := m["sprint_id"]; ok {
		t.Error("nil sprint_id should not be in map")
	}
}

func TestUpdateIssueRequest_toMap_clearSprintID(t *testing.T) {
	req := UpdateIssueRequest{ClearSprintID: true}
	m := req.toMap()
	v, ok := m["sprint_id"]
	if !ok {
		t.Fatal("expected sprint_id in map")
	}
	if v != nil {
		t.Errorf("expected nil for cleared sprint_id, got %v", v)
	}
}

func TestUpdateIssueRequest_toMap_clearParentID(t *testing.T) {
	req := UpdateIssueRequest{ClearParentID: true}
	m := req.toMap()
	v, ok := m["parent_id"]
	if !ok {
		t.Fatal("expected parent_id in map")
	}
	if v != nil {
		t.Errorf("expected nil for cleared parent_id, got %v", v)
	}
}

func TestUpdateIssueRequest_toMap_clearAssignees(t *testing.T) {
	req := UpdateIssueRequest{ClearAssignees: true}
	m := req.toMap()
	ids, ok := m["assignee_ids"].([]string)
	if !ok {
		t.Fatal("expected assignee_ids as []string")
	}
	if len(ids) != 0 {
		t.Errorf("expected empty slice, got %v", ids)
	}
}

func TestUpdateIssueRequest_toMap_clearLabels(t *testing.T) {
	req := UpdateIssueRequest{ClearLabels: true}
	m := req.toMap()
	ids, ok := m["label_ids"].([]string)
	if !ok {
		t.Fatal("expected label_ids as []string")
	}
	if len(ids) != 0 {
		t.Errorf("expected empty slice, got %v", ids)
	}
}

func TestUpdateIssueRequest_toMap_clearOwnerID(t *testing.T) {
	req := UpdateIssueRequest{ClearOwnerID: true}
	m := req.toMap()
	v, ok := m["owner_id"]
	if !ok {
		t.Fatal("expected owner_id in map")
	}
	if v != nil {
		t.Errorf("expected nil for cleared owner_id, got %v", v)
	}
}

func TestUpdateIssueRequest_toMap_setOverridesClear(t *testing.T) {
	// If both SprintID and ClearSprintID are set, ClearSprintID wins
	// (it's checked first in the if/else chain)
	req := UpdateIssueRequest{SprintID: ptr("uuid"), ClearSprintID: true}
	m := req.toMap()
	if m["sprint_id"] != nil {
		t.Errorf("ClearSprintID should produce nil, got %v", m["sprint_id"])
	}
}

func TestUpdateIssueRequest_toMap_assigneeIDs(t *testing.T) {
	req := UpdateIssueRequest{AssigneeIDs: []string{"a", "b"}}
	m := req.toMap()
	ids := m["assignee_ids"].([]string)
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Errorf("unexpected assignee_ids: %v", ids)
	}
}

func TestUpdateIssueRequest_toMap_labelIDs(t *testing.T) {
	req := UpdateIssueRequest{LabelIDs: []string{"l1"}}
	m := req.toMap()
	ids := m["label_ids"].([]string)
	if len(ids) != 1 || ids[0] != "l1" {
		t.Errorf("unexpected label_ids: %v", ids)
	}
}

func TestUpdateSprintRequest_toMap_allFields(t *testing.T) {
	req := UpdateSprintRequest{
		Name:                     ptr("Sprint 1"),
		Goal:                     ptr("Ship it"),
		StartDate:                ptr("2026-01-01T00:00:00Z"),
		EndDate:                  ptr("2026-01-14T00:00:00Z"),
		Status:                   ptr("completed"),
		Force:                    true,
		MoveOpenIssuesToSprintID: ptr("next-sprint-uuid"),
	}
	m := req.toMap()
	if m["name"] != "Sprint 1" {
		t.Errorf("expected name, got %v", m["name"])
	}
	if m["goal"] != "Ship it" {
		t.Errorf("expected goal, got %v", m["goal"])
	}
	if m["status"] != "completed" {
		t.Errorf("expected status, got %v", m["status"])
	}
	if m["force"] != true {
		t.Errorf("expected force=true, got %v", m["force"])
	}
	if m["move_open_issues_to_sprint_id"] != "next-sprint-uuid" {
		t.Errorf("expected move id, got %v", m["move_open_issues_to_sprint_id"])
	}
}

func TestUpdateSprintRequest_toMap_omitsNilAndFalseForce(t *testing.T) {
	req := UpdateSprintRequest{Name: ptr("Only Name")}
	m := req.toMap()
	if _, ok := m["force"]; ok {
		t.Error("force=false should not be in map")
	}
	if _, ok := m["status"]; ok {
		t.Error("nil status should not be in map")
	}
}

func TestUpdateProjectRequest_toMap_allFields(t *testing.T) {
	req := UpdateProjectRequest{
		Name:               ptr("New Name"),
		Description:        ptr("desc"),
		Archived:           boolPtr(true),
		WipLimitInProgress: intPtr(5),
		WipLimitInReview:   intPtr(3),
	}
	m := req.toMap()
	if m["name"] != "New Name" {
		t.Errorf("expected name, got %v", m["name"])
	}
	if m["archived"] != true {
		t.Errorf("expected archived=true, got %v", m["archived"])
	}
	if m["wip_limit_in_progress"] != 5 {
		t.Errorf("expected wip=5, got %v", m["wip_limit_in_progress"])
	}
}

func TestUpdateProjectRequest_toMap_wipLimitClear(t *testing.T) {
	req := UpdateProjectRequest{WipLimitInProgress: intPtr(-1)}
	m := req.toMap()
	v, ok := m["wip_limit_in_progress"]
	if !ok {
		t.Fatal("expected wip_limit_in_progress in map")
	}
	if v != nil {
		t.Errorf("expected nil for -1 wip limit, got %v", v)
	}
}

func TestUpdateMilestoneRequest_toMap_allFields(t *testing.T) {
	req := UpdateMilestoneRequest{
		Title:       ptr("v1.0"),
		Description: ptr("First release"),
		TargetDate:  ptr("2026-06-01T00:00:00Z"),
		Close:       boolPtr(true),
	}
	m := req.toMap()
	if m["title"] != "v1.0" {
		t.Errorf("expected title, got %v", m["title"])
	}
	if m["close"] != true {
		t.Errorf("expected close=true, got %v", m["close"])
	}
}

func TestUpdateMilestoneRequest_toMap_reopenMilestone(t *testing.T) {
	req := UpdateMilestoneRequest{Close: boolPtr(false)}
	m := req.toMap()
	if m["close"] != false {
		t.Errorf("expected close=false, got %v", m["close"])
	}
}

func TestBatchUpdateIssuesRequest_toMap_includesNumbers(t *testing.T) {
	status := "in_progress"
	req := BatchUpdateIssuesRequest{
		Numbers: []int{1, 2, 3},
		Status:  &status,
	}
	m := req.toMap()
	nums := m["numbers"].([]int)
	if len(nums) != 3 {
		t.Errorf("expected 3 numbers, got %d", len(nums))
	}
	if m["status"] != "in_progress" {
		t.Errorf("expected status, got %v", m["status"])
	}
}

func TestBatchUpdateIssuesRequest_toMap_clearSprintID(t *testing.T) {
	req := BatchUpdateIssuesRequest{
		Numbers:       []int{1},
		ClearSprintID: true,
	}
	m := req.toMap()
	if m["clear_sprint_id"] != true {
		t.Errorf("expected clear_sprint_id=true, got %v", m["clear_sprint_id"])
	}
	if _, ok := m["sprint_id"]; ok {
		t.Error("sprint_id should not be set when clearing")
	}
}
