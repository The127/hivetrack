package mcp

import (
	"testing"

	htclient "github.com/the127/hivetrack/client"
)

func TestFormatListIssues_empty(t *testing.T) {
	result := formatListIssues(nil, 0)
	if result != "No issues found." {
		t.Errorf("unexpected: %s", result)
	}
}

func TestFormatListIssues_withItems(t *testing.T) {
	items := []htclient.IssueSummary{
		{Number: 1, Type: "task", Title: "Fix bug", Status: "todo", Priority: "high", Estimate: "s"},
		{Number: 2, Type: "epic", Title: "Big Feature", Status: "in_progress", Priority: "none", OnHold: true},
	}
	result := formatListIssues(items, 2)
	if !contains(result, "2 issue(s)") {
		t.Errorf("expected count, got: %s", result)
	}
	if !contains(result, "#1") || !contains(result, "Fix bug") || !contains(result, "high") {
		t.Errorf("expected issue 1 details, got: %s", result)
	}
	if !contains(result, "◆") {
		t.Errorf("expected epic marker, got: %s", result)
	}
	if !contains(result, "ON HOLD") {
		t.Errorf("expected ON HOLD, got: %s", result)
	}
}

func TestFormatGetIssue_basic(t *testing.T) {
	desc := "Some description"
	issue := &htclient.IssueDetail{
		ID:       "uuid-1",
		Number:   42,
		Type:     "task",
		Title:    "Test Issue",
		Status:   "todo",
		Priority: "high",
		Estimate: "m",
		Triaged:  true,
		Description: &desc,
		Assignees: []htclient.UserInfo{{DisplayName: "Alice"}},
		Labels:    []htclient.LabelInfo{{Name: "bug"}},
		Links:     []htclient.IssueLinkInfo{{LinkType: "blocks", LinkedIssueNumber: 10}},
		Checklist: []htclient.ChecklistItem{{ID: "c1", Text: "Step 1", Done: true}},
	}
	result := formatGetIssue(issue)
	if !contains(result, "#42 Test Issue") {
		t.Errorf("expected title, got: %s", result)
	}
	if !contains(result, "Priority: high") {
		t.Errorf("expected priority, got: %s", result)
	}
	if !contains(result, "Alice") {
		t.Errorf("expected assignee, got: %s", result)
	}
	if !contains(result, "bug") {
		t.Errorf("expected label, got: %s", result)
	}
	if !contains(result, "blocks #10") {
		t.Errorf("expected link, got: %s", result)
	}
	if !contains(result, "☑ Step 1") {
		t.Errorf("expected checklist, got: %s", result)
	}
	if !contains(result, "Some description") {
		t.Errorf("expected description, got: %s", result)
	}
}

func TestFormatGetIssue_untriaged(t *testing.T) {
	issue := &htclient.IssueDetail{Number: 1, Title: "New", Type: "task", Status: "todo", Triaged: false}
	result := formatGetIssue(issue)
	if !contains(result, "Untriaged") {
		t.Errorf("expected untriaged marker, got: %s", result)
	}
}

func TestFormatGetIssue_epicMarker(t *testing.T) {
	issue := &htclient.IssueDetail{Number: 1, Title: "Epic", Type: "epic", Status: "todo", Triaged: true}
	result := formatGetIssue(issue)
	if !contains(result, "◆ epic") {
		t.Errorf("expected epic marker, got: %s", result)
	}
}

func TestFormatListSprints_empty(t *testing.T) {
	if formatListSprints(nil) != "No sprints found." {
		t.Error("expected empty message")
	}
}

func TestFormatListSprints_withData(t *testing.T) {
	sprints := []htclient.Sprint{{ID: "s1", Name: "Sprint 1", Status: "active", Goal: "Ship it"}}
	result := formatListSprints(sprints)
	if !contains(result, "Sprint 1") || !contains(result, "active") || !contains(result, "Ship it") {
		t.Errorf("unexpected: %s", result)
	}
}

func TestFormatListProjects_empty(t *testing.T) {
	if formatListProjects(nil) != "No projects found." {
		t.Error("expected empty message")
	}
}

func TestFormatListComments_empty(t *testing.T) {
	if formatListComments(nil, 0) != "No comments." {
		t.Error("expected empty message")
	}
}

func TestFormatListComments_withID(t *testing.T) {
	comments := []htclient.Comment{{ID: "c-uuid", AuthorName: "Alice", Body: "Hello", CreatedAt: "2026-01-01"}}
	result := formatListComments(comments, 1)
	if !contains(result, "c-uuid") {
		t.Errorf("expected comment ID, got: %s", result)
	}
}

func TestFormatListLabels_empty(t *testing.T) {
	if formatListLabels(nil) != "No labels found." {
		t.Error("expected empty message")
	}
}

func TestFormatListMilestones_empty(t *testing.T) {
	if formatListMilestones(nil) != "No milestones found." {
		t.Error("expected empty message")
	}
}

func TestFormatListMilestones_withProgress(t *testing.T) {
	target := "2026-06-01"
	ms := []htclient.Milestone{{ID: "m1", Title: "v1.0", TargetDate: &target, IssueCount: 10, ClosedIssueCount: 7}}
	result := formatListMilestones(ms)
	if !contains(result, "7/10") {
		t.Errorf("expected progress, got: %s", result)
	}
	if !contains(result, "2026-06-01") {
		t.Errorf("expected target date, got: %s", result)
	}
}

func TestFormatCurrentUser_admin(t *testing.T) {
	user := &htclient.User{ID: "u1", Email: "admin@test.com", IsAdmin: true}
	result := formatCurrentUser(user)
	if !contains(result, "[admin]") {
		t.Errorf("expected admin marker, got: %s", result)
	}
}

func TestFormatSprintBurndown(t *testing.T) {
	bd := &htclient.BurndownData{
		Total: 10, StartRemaining: 10, EndRemaining: 3,
		Points: []htclient.BurndownPoint{{Date: "2026-01-01", Remaining: 8}},
	}
	result := formatSprintBurndown(bd)
	if !contains(result, "Total: 10") || !contains(result, "2026-01-01") || !contains(result, "8") {
		t.Errorf("unexpected: %s", result)
	}
}

func TestFormatSplitIssue(t *testing.T) {
	result := formatSplitIssue(&htclient.SplitIssueResult{
		NewIssues: []htclient.CreateIssueResult{{Number: 5}, {Number: 6}},
	})
	if !contains(result, "#5") || !contains(result, "#6") || !contains(result, "2 issues") {
		t.Errorf("unexpected: %s", result)
	}
}

func TestFormatTriageIssue_withExtras(t *testing.T) {
	result := formatTriageIssue(5, "todo", map[string]any{
		"sprint_id": "s1",
		"priority":  "high",
		"estimate":  "m",
	})
	if !contains(result, "priority=high") || !contains(result, "estimate=M") || !contains(result, "sprint") {
		t.Errorf("unexpected: %s", result)
	}
}

func TestFormatUpdateIssue_OnHold(t *testing.T) {
	result := formatUpdateIssue(7, map[string]any{
		"on_hold":     "true",
		"hold_reason": "waiting_on_customer",
	})
	if !contains(result, "on hold") || !contains(result, "waiting_on_customer") {
		t.Errorf("unexpected: %s", result)
	}
}

func TestFormatUpdateIssue_ClearHold(t *testing.T) {
	result := formatUpdateIssue(7, map[string]any{
		"on_hold": "false",
	})
	if !contains(result, "hold cleared") {
		t.Errorf("unexpected: %s", result)
	}
}

func TestFormatCreateIssue(t *testing.T) {
	result := formatCreateIssue(&htclient.CreateIssueResult{Number: 42}, map[string]any{
		"title":    "New Thing",
		"type":     "task",
		"priority": "high",
	})
	if !contains(result, "#42") || !contains(result, "New Thing") || !contains(result, "high") {
		t.Errorf("unexpected: %s", result)
	}
}
