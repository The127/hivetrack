package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

// AddChecklistItemCommand adds a new checklist item to an issue.
type AddChecklistItemCommand struct {
	IssueID uuid.UUID
	Text    string
}

type AddChecklistItemResult struct {
	ID uuid.UUID
}

func HandleAddChecklistItem(ctx context.Context, cmd AddChecklistItemCommand) (*AddChecklistItemResult, error) {
	db := repositories.GetDbContext(ctx)

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	itemID := uuid.New()
	checklist := append(issue.GetChecklist(), models.ChecklistItem{
		ID:   itemID,
		Text: cmd.Text,
		Done: false,
	})
	issue.SetChecklist(checklist)
	issue.SetUpdatedAt(time.Now())

	db.Issues().Update(issue)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving issue: %w", err)
	}
	return &AddChecklistItemResult{ID: itemID}, nil
}

// UpdateChecklistItemCommand updates an existing checklist item's text and/or done status.
type UpdateChecklistItemCommand struct {
	IssueID uuid.UUID
	ItemID  uuid.UUID
	Text    *string
	Done    *bool
}

type UpdateChecklistItemResult struct{}

func HandleUpdateChecklistItem(ctx context.Context, cmd UpdateChecklistItemCommand) (*UpdateChecklistItemResult, error) {
	db := repositories.GetDbContext(ctx)

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	checklist := issue.GetChecklist()
	found := false
	for i, item := range checklist {
		if item.ID == cmd.ItemID {
			if cmd.Text != nil {
				checklist[i].Text = *cmd.Text
			}
			if cmd.Done != nil {
				checklist[i].Done = *cmd.Done
			}
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("checklist item %s: %w", cmd.ItemID, models.ErrNotFound)
	}

	issue.SetChecklist(checklist)
	issue.SetUpdatedAt(time.Now())

	db.Issues().Update(issue)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving issue: %w", err)
	}
	return &UpdateChecklistItemResult{}, nil
}

// RemoveChecklistItemCommand removes a checklist item from an issue.
type RemoveChecklistItemCommand struct {
	IssueID uuid.UUID
	ItemID  uuid.UUID
}

type RemoveChecklistItemResult struct{}

func HandleRemoveChecklistItem(ctx context.Context, cmd RemoveChecklistItemCommand) (*RemoveChecklistItemResult, error) {
	db := repositories.GetDbContext(ctx)

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	checklist := issue.GetChecklist()
	filtered := make([]models.ChecklistItem, 0, len(checklist))
	found := false
	for _, item := range checklist {
		if item.ID == cmd.ItemID {
			found = true
			continue
		}
		filtered = append(filtered, item)
	}
	if !found {
		return nil, fmt.Errorf("checklist item %s: %w", cmd.ItemID, models.ErrNotFound)
	}

	issue.SetChecklist(filtered)
	issue.SetUpdatedAt(time.Now())

	db.Issues().Update(issue)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving issue: %w", err)
	}
	return &RemoveChecklistItemResult{}, nil
}
