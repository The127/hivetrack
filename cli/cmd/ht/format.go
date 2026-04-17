package main

import (
	"fmt"
	"strings"

	htclient "github.com/the127/hivetrack/client"
)

func formatProjects(projects []htclient.ProjectSummary) string {
	if len(projects) == 0 {
		return "No projects found."
	}
	var sb strings.Builder
	for _, p := range projects {
		fmt.Fprintf(&sb, "• %-20s  %-12s  %s\n", p.Name, p.Slug, p.Archetype)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func formatIssueList(items []htclient.IssueSummary, total int) string {
	if len(items) == 0 {
		return "No issues found."
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d issue(s):\n", total)
	for _, item := range items {
		typeMarker := "  "
		if item.Type == htclient.IssueTypeEpic {
			typeMarker = "◆ "
		}

		meta := []string{string(item.Status)}
		if item.Priority != htclient.IssuePriorityNone && item.Priority != "" {
			meta = append(meta, string(item.Priority))
		}
		if item.Estimate != htclient.IssueEstimateNone && item.Estimate != "" {
			meta = append(meta, strings.ToUpper(string(item.Estimate)))
		}
		if !item.Triaged {
			meta = append(meta, "untriaged")
		}
		if item.OnHold {
			meta = append(meta, "ON HOLD")
		}

		assignees := ""
		if len(item.Assignees) > 0 {
			names := make([]string, 0, len(item.Assignees))
			for _, a := range item.Assignees {
				names = append(names, a.DisplayName)
			}
			assignees = "  → " + strings.Join(names, ", ")
		}

		fmt.Fprintf(&sb, "%s#%-4d  %-50s  (%s)%s\n",
			typeMarker, item.Number, item.Title,
			strings.Join(meta, ", "), assignees)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func formatIssue(issue *htclient.IssueDetail) string {
	var sb strings.Builder

	typeLabel := string(issue.Type)
	if issue.Type == htclient.IssueTypeEpic {
		typeLabel = "epic ◆"
	}
	fmt.Fprintf(&sb, "#%d  %s  [%s]\n", issue.Number, issue.Title, typeLabel)
	fmt.Fprintf(&sb, "ID:       %s\n", issue.ID)
	fmt.Fprintf(&sb, "Status:   %s", issue.Status)
	if issue.OnHold {
		sb.WriteString(" (ON HOLD)")
		if issue.HoldReason != nil {
			fmt.Fprintf(&sb, " — %s", *issue.HoldReason)
		}
	}
	sb.WriteString("\n")

	if issue.Priority != htclient.IssuePriorityNone {
		fmt.Fprintf(&sb, "Priority: %s\n", issue.Priority)
	}
	if issue.Estimate != htclient.IssueEstimateNone {
		fmt.Fprintf(&sb, "Estimate: %s\n", strings.ToUpper(string(issue.Estimate)))
	}
	if !issue.Triaged {
		sb.WriteString("⚠  Untriaged\n")
	}
	if !issue.Refined {
		sb.WriteString("⚠  Not refined\n")
	}

	if len(issue.Assignees) > 0 {
		names := make([]string, 0, len(issue.Assignees))
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
		names := make([]string, 0, len(issue.Labels))
		for _, l := range issue.Labels {
			names = append(names, l.Name)
		}
		fmt.Fprintf(&sb, "Labels:   %s\n", strings.Join(names, ", "))
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
			fmt.Fprintf(&sb, "  %s %s\n", check, item.Text)
		}
		done := 0
		for _, item := range issue.Checklist {
			if item.Done {
				done++
			}
		}
		fmt.Fprintf(&sb, "  %d/%d done\n", done, len(issue.Checklist))
	}

	return strings.TrimRight(sb.String(), "\n")
}

func formatSprints(sprints []htclient.Sprint) string {
	if len(sprints) == 0 {
		return "No sprints found."
	}
	var sb strings.Builder
	for _, s := range sprints {
		goal := s.Goal
		if goal == "" {
			goal = "(no goal)"
		}
		fmt.Fprintf(&sb, "• %-30s  [%s]  %s – %s\n  id: %s\n  goal: %s\n",
			s.Name, s.Status, s.StartDate, s.EndDate, s.ID, goal)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func formatMilestones(milestones []htclient.Milestone) string {
	if len(milestones) == 0 {
		return "No milestones found."
	}
	var sb strings.Builder
	for _, m := range milestones {
		status := "open"
		if m.ClosedAt != nil {
			status = "closed"
		}
		target := ""
		if m.TargetDate != nil {
			td := *m.TargetDate
			if len(td) > 10 {
				td = td[:10]
			}
			target = "  target: " + td
		}
		fmt.Fprintf(&sb, "• %-30s  [%s%s]  %d/%d done\n  id: %s\n",
			m.Title, status, target, m.ClosedIssueCount, m.IssueCount, m.ID)
	}
	return strings.TrimRight(sb.String(), "\n")
}
