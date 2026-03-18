package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatCreateIssue formats a create_issue response for human readability.
func formatCreateIssue(data json.RawMessage, args map[string]any) string {
	var resp struct {
		ID     string `json:"ID"`
		Number int    `json:"Number"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return string(data)
	}

	issueType := stringOr(args, "type", "task")
	title, _ := args["title"].(string)
	priority := stringOr(args, "priority", "")
	estimate := stringOr(args, "estimate", "")

	var parts []string
	if priority != "" {
		parts = append(parts, priority)
	}
	if estimate != "" {
		parts = append(parts, strings.ToUpper(estimate))
	}

	meta := ""
	if len(parts) > 0 {
		meta = ", " + strings.Join(parts, ", ")
	}

	return fmt.Sprintf("Created %s #%d: %q (%s%s)", issueType, resp.Number, title, issueType, meta)
}

// formatTriageIssue formats a triage response.
func formatTriageIssue(number int, status string, args map[string]any) string {
	msg := fmt.Sprintf("Triaged #%d → %s", number, status)
	if sprintID, ok := args["sprint_id"].(string); ok && sprintID != "" {
		msg += " (assigned to sprint)"
	}
	return msg
}

// formatUpdateIssue formats an update_issue response.
func formatUpdateIssue(number int, args map[string]any) string {
	var changes []string
	for _, key := range []string{"title", "status", "priority", "estimate", "description"} {
		if v, ok := args[key].(string); ok && v != "" {
			changes = append(changes, fmt.Sprintf("%s → %s", key, v))
		}
	}
	for _, key := range []string{"sprint_id", "milestone_id", "parent_id"} {
		if v, ok := args[key].(string); ok {
			if v == "null" {
				changes = append(changes, fmt.Sprintf("%s cleared", strings.TrimSuffix(key, "_id")))
			} else if v != "" {
				changes = append(changes, fmt.Sprintf("%s set", strings.TrimSuffix(key, "_id")))
			}
		}
	}
	if v, ok := args["assignee_ids"].(string); ok {
		if v == "null" {
			changes = append(changes, "assignees cleared")
		} else if v != "" {
			changes = append(changes, "assignees updated")
		}
	}
	return fmt.Sprintf("Updated #%d: %s", number, strings.Join(changes, ", "))
}

// formatCreateSprint formats a create_sprint response.
func formatCreateSprint(data json.RawMessage, name string) string {
	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return string(data)
	}
	return fmt.Sprintf("Created sprint %q (id: %s)", name, resp.ID)
}

// formatUpdateSprint formats an update_sprint response.
func formatUpdateSprint(args map[string]any) string {
	var changes []string
	for _, key := range []string{"name", "goal", "status", "start_date", "end_date"} {
		if v, ok := args[key].(string); ok && v != "" {
			changes = append(changes, fmt.Sprintf("%s → %s", key, v))
		}
	}
	return fmt.Sprintf("Updated sprint: %s", strings.Join(changes, ", "))
}

// formatCreateProject formats a create_project response.
func formatCreateProject(data json.RawMessage, slug, name, archetype string) string {
	return fmt.Sprintf("Created project %q (%s, %s)", name, slug, archetype)
}

// formatListIssues formats a list_issues response as a compact table.
func formatListIssues(data json.RawMessage) string {
	var resp struct {
		Items []struct {
			Number    int      `json:"number"`
			Type      string   `json:"type"`
			Title     string   `json:"title"`
			Status    string   `json:"status"`
			Priority  string   `json:"priority"`
			Estimate  string   `json:"estimate"`
			Triaged   bool     `json:"triaged"`
			ParentID  *string  `json:"parent_id"`
			SprintID  *string  `json:"sprint_id"`
			OnHold    bool     `json:"on_hold"`
			Labels    []string `json:"labels"`
			Assignees []struct {
				DisplayName string `json:"display_name"`
			} `json:"assignees"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return string(data)
	}

	if len(resp.Items) == 0 {
		return "No issues found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d issue(s):\n\n", resp.Total))
	for _, item := range resp.Items {
		marker := "  "
		if item.Type == "epic" {
			marker = "◆ "
		}

		meta := []string{item.Status}
		if item.Priority != "" && item.Priority != "none" {
			meta = append(meta, item.Priority)
		}
		if item.Estimate != "" && item.Estimate != "none" {
			meta = append(meta, strings.ToUpper(item.Estimate))
		}
		if !item.Triaged {
			meta = append(meta, "untriaged")
		}
		if item.OnHold {
			meta = append(meta, "ON HOLD")
		}

		assigneeStr := ""
		if len(item.Assignees) > 0 {
			var names []string
			for _, a := range item.Assignees {
				names = append(names, a.DisplayName)
			}
			assigneeStr = " → " + strings.Join(names, ", ")
		}

		sb.WriteString(fmt.Sprintf("%s#%-4d %-50s (%s)%s\n", marker, item.Number, item.Title, strings.Join(meta, ", "), assigneeStr))
	}
	return sb.String()
}

// formatListSprints formats a list_sprints response.
func formatListSprints(data json.RawMessage) string {
	var resp struct {
		Sprints []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Goal      string `json:"goal"`
			Status    string `json:"status"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
		} `json:"sprints"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return string(data)
	}

	if len(resp.Sprints) == 0 {
		return "No sprints found."
	}

	var sb strings.Builder
	for _, s := range resp.Sprints {
		sb.WriteString(fmt.Sprintf("• %s [%s] — %s\n  id: %s\n", s.Name, s.Status, s.Goal, s.ID))
	}
	return sb.String()
}

// formatListProjects formats a list_projects response.
func formatListProjects(data json.RawMessage) string {
	var resp struct {
		Items []struct {
			Slug      string `json:"slug"`
			Name      string `json:"name"`
			Archetype string `json:"archetype"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return string(data)
	}

	if len(resp.Items) == 0 {
		return "No projects found."
	}

	var sb strings.Builder
	for _, p := range resp.Items {
		sb.WriteString(fmt.Sprintf("• %s (%s, %s)\n", p.Name, p.Slug, p.Archetype))
	}
	return sb.String()
}

// formatListComments formats a list_comments response.
func formatListComments(data json.RawMessage) string {
	var resp struct {
		Items []struct {
			ID          string `json:"id"`
			AuthorName  string `json:"author_name"`
			AuthorEmail string `json:"author_email"`
			Body        string `json:"body"`
			CreatedAt   string `json:"created_at"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return string(data)
	}

	if len(resp.Items) == 0 {
		return "No comments."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d comment(s):\n\n", resp.Total))
	for _, c := range resp.Items {
		author := c.AuthorName
		if author == "" {
			author = c.AuthorEmail
		}
		if author == "" {
			author = "unknown"
		}
		sb.WriteString(fmt.Sprintf("— %s (%s):\n%s\n\n", author, c.CreatedAt, c.Body))
	}
	return sb.String()
}

// formatGetIssue formats a get_issue response with full details.
func formatGetIssue(data json.RawMessage) string {
	var issue struct {
		ID          string  `json:"id"`
		Number      int     `json:"number"`
		Type        string  `json:"type"`
		Title       string  `json:"title"`
		Status      string  `json:"status"`
		Priority    string  `json:"priority"`
		Estimate    string  `json:"estimate"`
		Description string  `json:"description"`
		Triaged     bool    `json:"triaged"`
		OnHold      bool    `json:"on_hold"`
		ParentID    *string `json:"parent_id"`
		SprintID    *string `json:"sprint_id"`
		Assignees   []struct {
			DisplayName string `json:"display_name"`
			Email       string `json:"email"`
		} `json:"assignees"`
		Checklist []struct {
			ID   string `json:"id"`
			Text string `json:"text"`
			Done bool   `json:"done"`
		} `json:"checklist"`
	}
	if err := json.Unmarshal(data, &issue); err != nil {
		return string(data)
	}

	var sb strings.Builder
	typeLabel := issue.Type
	if issue.Type == "epic" {
		typeLabel = "◆ epic"
	}
	sb.WriteString(fmt.Sprintf("#%d %s [%s]\n", issue.Number, issue.Title, typeLabel))
	sb.WriteString(fmt.Sprintf("ID: %s\n", issue.ID))
	sb.WriteString(fmt.Sprintf("Status: %s", issue.Status))
	if issue.OnHold {
		sb.WriteString(" (ON HOLD)")
	}
	sb.WriteString("\n")

	if issue.Priority != "" && issue.Priority != "none" {
		sb.WriteString(fmt.Sprintf("Priority: %s\n", issue.Priority))
	}
	if issue.Estimate != "" && issue.Estimate != "none" {
		sb.WriteString(fmt.Sprintf("Estimate: %s\n", strings.ToUpper(issue.Estimate)))
	}
	if !issue.Triaged {
		sb.WriteString("⚠ Untriaged\n")
	}

	if len(issue.Assignees) > 0 {
		var names []string
		for _, a := range issue.Assignees {
			if a.DisplayName != "" {
				names = append(names, a.DisplayName)
			} else {
				names = append(names, a.Email)
			}
		}
		sb.WriteString(fmt.Sprintf("Assignees: %s\n", strings.Join(names, ", ")))
	}

	if issue.Description != "" {
		sb.WriteString(fmt.Sprintf("\n%s\n", issue.Description))
	}

	if len(issue.Checklist) > 0 {
		sb.WriteString("\nChecklist:\n")
		for _, item := range issue.Checklist {
			check := "☐"
			if item.Done {
				check = "☑"
			}
			sb.WriteString(fmt.Sprintf("  %s %s  (id: %s)\n", check, item.Text, item.ID))
		}
	}

	return sb.String()
}
