package mcp

import (
	"fmt"
	"strings"

	htclient "github.com/the127/hivetrack/client"
)

// formatCreateIssue formats a create_issue result for human readability.
func formatCreateIssue(result *htclient.CreateIssueResult, args map[string]any) string {
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

	return fmt.Sprintf("Created %s #%d: %q (%s%s)", issueType, result.Number, title, issueType, meta)
}

// formatTriageIssue formats a triage response.
func formatTriageIssue(number int, status string, args map[string]any) string {
	msg := fmt.Sprintf("Triaged #%d → %s", number, status)
	if sprintID, ok := args["sprint_id"].(string); ok && sprintID != "" {
		msg += " (assigned to sprint)"
	}
	var extras []string
	if p, ok := args["priority"].(string); ok && p != "" {
		extras = append(extras, "priority="+p)
	}
	if e, ok := args["estimate"].(string); ok && e != "" {
		extras = append(extras, "estimate="+strings.ToUpper(e))
	}
	if len(extras) > 0 {
		msg += ", " + strings.Join(extras, ", ")
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
	if v, ok := args["label_ids"].(string); ok {
		if v == "null" {
			changes = append(changes, "labels cleared")
		} else if v != "" {
			changes = append(changes, "labels updated")
		}
	}
	if v, ok := args["on_hold"].(string); ok && v != "" {
		if v == "true" {
			reason, _ := args["hold_reason"].(string)
			if reason != "" {
				changes = append(changes, "on hold ("+reason+")")
			} else {
				changes = append(changes, "on hold")
			}
		} else {
			changes = append(changes, "hold cleared")
		}
	}
	return fmt.Sprintf("Updated #%d: %s", number, strings.Join(changes, ", "))
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

// formatListIssues formats issue summaries as a compact table.
func formatListIssues(items []htclient.IssueSummary, total int) string {
	if len(items) == 0 {
		return "No issues found."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d issue(s):\n\n", total)
	for _, item := range items {
		marker := "  "
		if item.Type == htclient.IssueTypeEpic {
			marker = "◆ "
		}

		meta := []string{string(item.Status)}
		if item.Priority != "" && item.Priority != htclient.IssuePriorityNone {
			meta = append(meta, string(item.Priority))
		}
		if item.Estimate != "" && item.Estimate != htclient.IssueEstimateNone {
			meta = append(meta, strings.ToUpper(string(item.Estimate)))
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

		labelStr := ""
		if len(item.Labels) > 0 {
			var labelNames []string
			for _, l := range item.Labels {
				labelNames = append(labelNames, l.Name)
			}
			labelStr = " [" + strings.Join(labelNames, ", ") + "]"
		}

		fmt.Fprintf(&sb, "%s#%-4d %-50s (%s)%s%s\n", marker, item.Number, item.Title, strings.Join(meta, ", "), assigneeStr, labelStr)
	}
	return sb.String()
}

// formatGetIssue formats an issue detail with full information.
func formatGetIssue(issue *htclient.IssueDetail) string {
	var sb strings.Builder
	typeLabel := string(issue.Type)
	if issue.Type == htclient.IssueTypeEpic {
		typeLabel = "◆ epic"
	}
	fmt.Fprintf(&sb, "#%d %s [%s]\n", issue.Number, issue.Title, typeLabel)
	fmt.Fprintf(&sb, "ID: %s\n", issue.ID)
	fmt.Fprintf(&sb, "Status: %s", issue.Status)
	if issue.OnHold {
		sb.WriteString(" (ON HOLD)")
	}
	sb.WriteString("\n")

	if issue.Priority != "" && issue.Priority != htclient.IssuePriorityNone {
		fmt.Fprintf(&sb, "Priority: %s\n", issue.Priority)
	}
	if issue.Estimate != "" && issue.Estimate != htclient.IssueEstimateNone {
		fmt.Fprintf(&sb, "Estimate: %s\n", strings.ToUpper(string(issue.Estimate)))
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
		fmt.Fprintf(&sb, "Assignees: %s\n", strings.Join(names, ", "))
	}

	if len(issue.Labels) > 0 {
		var labelNames []string
		for _, l := range issue.Labels {
			labelNames = append(labelNames, l.Name)
		}
		fmt.Fprintf(&sb, "Labels: %s\n", strings.Join(labelNames, ", "))
	}

	if issue.Description != nil && *issue.Description != "" {
		fmt.Fprintf(&sb, "\n%s\n", *issue.Description)
	}

	if len(issue.Links) > 0 {
		sb.WriteString("\nLinks:\n")
		for _, l := range issue.Links {
			fmt.Fprintf(&sb, "  %s #%d\n", l.LinkType, l.LinkedIssueNumber)
		}
	}

	if len(issue.Checklist) > 0 {
		sb.WriteString("\nChecklist:\n")
		for _, item := range issue.Checklist {
			check := "☐"
			if item.Done {
				check = "☑"
			}
			fmt.Fprintf(&sb, "  %s %s  (id: %s)\n", check, item.Text, item.ID)
		}
	}

	return sb.String()
}

// formatListSprints formats sprint summaries.
func formatListSprints(sprints []htclient.Sprint) string {
	if len(sprints) == 0 {
		return "No sprints found."
	}

	var sb strings.Builder
	for _, s := range sprints {
		fmt.Fprintf(&sb, "• %s [%s] — %s\n  id: %s\n", s.Name, s.Status, s.Goal, s.ID)
	}
	return sb.String()
}

// formatListProjects formats project summaries.
func formatListProjects(projects []htclient.ProjectSummary) string {
	if len(projects) == 0 {
		return "No projects found."
	}

	var sb strings.Builder
	for _, p := range projects {
		fmt.Fprintf(&sb, "• %s (%s, %s)\n", p.Name, p.Slug, p.Archetype)
	}
	return sb.String()
}

// formatListComments formats comment list.
func formatListComments(comments []htclient.Comment, total int) string {
	if len(comments) == 0 {
		return "No comments."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d comment(s):\n\n", total)
	for _, c := range comments {
		author := c.AuthorName
		if author == "" {
			author = c.AuthorEmail
		}
		if author == "" {
			author = "unknown"
		}
		fmt.Fprintf(&sb, "— %s (%s) [id: %s]:\n%s\n\n", author, c.CreatedAt, c.ID, c.Body)
	}
	return sb.String()
}

// formatListLabels formats label list.
func formatListLabels(labels []htclient.LabelInfo) string {
	if len(labels) == 0 {
		return "No labels found."
	}

	var sb strings.Builder
	for _, l := range labels {
		fmt.Fprintf(&sb, "• %s (%s) id: %s\n", l.Name, l.Color, l.ID)
	}
	return sb.String()
}

// formatListMilestones formats milestone list.
func formatListMilestones(milestones []htclient.Milestone) string {
	if len(milestones) == 0 {
		return "No milestones found."
	}

	var sb strings.Builder
	for _, m := range milestones {
		status := "open"
		if m.ClosedAt != nil {
			status = "closed"
		}

		progress := fmt.Sprintf("%d/%d issues done", m.ClosedIssueCount, m.IssueCount)

		target := ""
		if m.TargetDate != nil {
			td := *m.TargetDate
			if len(td) > 10 {
				td = td[:10]
			}
			target = fmt.Sprintf(", target: %s", td)
		}

		fmt.Fprintf(&sb, "• %s [%s%s] — %s\n  id: %s\n", m.Title, status, target, progress, m.ID)
	}
	return sb.String()
}

// formatCurrentUser formats a user response.
func formatCurrentUser(user *htclient.User) string {
	adminStr := ""
	if user.IsAdmin {
		adminStr = " [admin]"
	}
	return fmt.Sprintf("%s (%s)%s", user.Email, user.ID, adminStr)
}

// formatSplitIssue formats a split result.
func formatSplitIssue(result *htclient.SplitIssueResult) string {
	nums := make([]string, 0, len(result.NewIssues))
	for _, issue := range result.NewIssues {
		nums = append(nums, fmt.Sprintf("#%d", issue.Number))
	}
	return fmt.Sprintf("Split into %d issues: %s", len(result.NewIssues), strings.Join(nums, ", "))
}

// formatSprintBurndown formats burndown chart data as a table.
func formatSprintBurndown(data *htclient.BurndownData) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Total: %d | Start: %d | End: %d\n\n", data.Total, data.StartRemaining, data.EndRemaining)
	sb.WriteString("Date        | Remaining\n")
	sb.WriteString("------------|----------\n")
	for _, p := range data.Points {
		fmt.Fprintf(&sb, "%-12s| %d\n", p.Date, p.Remaining)
	}
	return sb.String()
}

// formatCreateSprint formats a create sprint confirmation.
func formatCreateSprint(id, name string) string {
	return fmt.Sprintf("Created sprint %q (id: %s)", name, id)
}

// formatCreateLabel formats a create label confirmation.
func formatCreateLabel(id, name, color string) string {
	return fmt.Sprintf("Created label %q (%s, id: %s)", name, color, id)
}

// formatCreateProject formats a create project confirmation.
func formatCreateProject(slug, name, archetype string) string {
	return fmt.Sprintf("Created project %q (%s, %s)", name, slug, archetype)
}
