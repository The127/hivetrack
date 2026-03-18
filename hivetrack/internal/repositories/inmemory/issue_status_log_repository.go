package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type statusLogEntry struct {
	issueID   uuid.UUID
	status    string
	changedAt time.Time
}

type IssueStatusLogRepository struct {
	entries   []statusLogEntry
	issueRepo *IssueRepository
}

func NewIssueStatusLogRepository(issueRepo *IssueRepository) *IssueStatusLogRepository {
	return &IssueStatusLogRepository{issueRepo: issueRepo}
}

func (r *IssueStatusLogRepository) Insert(_ context.Context, issueID uuid.UUID, status string, changedAt time.Time) error {
	r.entries = append(r.entries, statusLogEntry{issueID: issueID, status: status, changedAt: changedAt})
	return nil
}

func (r *IssueStatusLogRepository) GetBurndownPoints(
	_ context.Context,
	sprintID uuid.UUID,
	startDate, endDate time.Time,
	terminalStatuses []string,
) ([]repositories.BurndownPoint, error) {
	terminal := make(map[string]bool, len(terminalStatuses))
	for _, s := range terminalStatuses {
		terminal[s] = true
	}

	// Collect task issues in this sprint
	var sprintIssues []*models.Issue
	for _, issue := range r.issueRepo.byID {
		if issue.GetSprintID() != nil && *issue.GetSprintID() == sprintID &&
			issue.GetType() == models.IssueTypeTask {
			sprintIssues = append(sprintIssues, issue)
		}
	}

	// Cap end date at today
	today := time.Now().UTC().Truncate(24 * time.Hour)
	effectiveEnd := endDate.UTC().Truncate(24 * time.Hour)
	if effectiveEnd.After(today) {
		effectiveEnd = today
	}

	var points []repositories.BurndownPoint
	for d := startDate.UTC().Truncate(24 * time.Hour); !d.After(effectiveEnd); d = d.Add(24 * time.Hour) {
		remaining := 0
		for _, issue := range sprintIssues {
			if !hasTerminalEntryOnOrBefore(r.entries, issue.GetId(), terminal, d) {
				remaining++
			}
		}
		points = append(points, repositories.BurndownPoint{Date: d, Remaining: remaining})
	}
	return points, nil
}

func hasTerminalEntryOnOrBefore(entries []statusLogEntry, issueID uuid.UUID, terminal map[string]bool, day time.Time) bool {
	dayEnd := day.Add(24 * time.Hour)
	for _, e := range entries {
		if e.issueID == issueID && terminal[e.status] && e.changedAt.Before(dayEnd) {
			return true
		}
	}
	return false
}
