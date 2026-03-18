package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type CreateCommentCommand struct {
	IssueID  uuid.UUID
	AuthorID uuid.UUID
	Body     string
}

type CreateCommentResult struct {
	ID uuid.UUID `json:"id"`
}

func HandleCreateComment(ctx context.Context, cmd CreateCommentCommand) (*CreateCommentResult, error) {
	db := repositories.GetDbContext(ctx)

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	comment := models.NewComment(cmd.IssueID, &cmd.AuthorID, nil, nil, cmd.Body)
	db.Comments().Insert(comment)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving comment: %w", err)
	}

	return &CreateCommentResult{ID: comment.GetId()}, nil
}
