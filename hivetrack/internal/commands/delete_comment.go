package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type DeleteCommentCommand struct {
	CommentID uuid.UUID
}

type DeleteCommentResult struct{}

func HandleDeleteComment(ctx context.Context, cmd DeleteCommentCommand) (*DeleteCommentResult, error) {
	db := repositories.GetDbContext(ctx)

	comment, err := db.Comments().GetByID(ctx, cmd.CommentID)
	if err != nil {
		return nil, fmt.Errorf("getting comment: %w", err)
	}
	if comment == nil {
		return nil, fmt.Errorf("comment %s: %w", cmd.CommentID, models.ErrNotFound)
	}

	db.Comments().Delete(comment)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("deleting comment: %w", err)
	}

	return &DeleteCommentResult{}, nil
}
