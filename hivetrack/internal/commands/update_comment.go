package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type UpdateCommentCommand struct {
	CommentID uuid.UUID
	Body      string
}

type UpdateCommentResult struct{}

func HandleUpdateComment(ctx context.Context, cmd UpdateCommentCommand) (*UpdateCommentResult, error) {
	db := repositories.GetDbContext(ctx)

	comment, err := db.Comments().GetByID(ctx, cmd.CommentID)
	if err != nil {
		return nil, fmt.Errorf("getting comment: %w", err)
	}
	if comment == nil {
		return nil, fmt.Errorf("comment %s: %w", cmd.CommentID, models.ErrNotFound)
	}

	comment.SetBody(cmd.Body)
	comment.SetUpdatedAt(time.Now())
	db.Comments().Update(comment)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("updating comment: %w", err)
	}

	return &UpdateCommentResult{}, nil
}
