package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetCommentsQuery struct {
	ProjectSlug string
	IssueNumber int
	Limit       int
	Offset      int
}

type CommentItem struct {
	ID          uuid.UUID  `json:"id"`
	AuthorID    *uuid.UUID `json:"author_id,omitempty"`
	AuthorName  string     `json:"author_name"`
	AuthorEmail string     `json:"author_email,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	Body        string     `json:"body"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type GetCommentsResult struct {
	Items  []CommentItem `json:"items"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

func HandleGetComments(ctx context.Context, q GetCommentsQuery) (*GetCommentsResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", q.ProjectSlug, models.ErrNotFound)
	}

	issue, err := db.Issues().GetByNumber(ctx, project.GetId(), q.IssueNumber)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue #%d: %w", q.IssueNumber, models.ErrNotFound)
	}

	comments, total, err := db.Comments().List(ctx, issue.GetId(), q.Limit, q.Offset)
	if err != nil {
		return nil, fmt.Errorf("listing comments: %w", err)
	}

	items := make([]CommentItem, 0, len(comments))
	for _, c := range comments {
		item := CommentItem{
			ID:        c.GetId(),
			AuthorID:  c.GetAuthorID(),
			Body:      c.GetBody(),
			CreatedAt: c.GetCreatedAt(),
			UpdatedAt: c.GetUpdatedAt(),
		}

		// Resolve author display info
		if c.GetAuthorID() != nil {
			user, err := db.Users().GetByID(ctx, *c.GetAuthorID())
			if err != nil {
				return nil, fmt.Errorf("resolving comment author: %w", err)
			}
			if user != nil {
				item.AuthorName = user.GetDisplayName()
				item.AuthorEmail = user.GetEmail()
				item.AvatarURL = user.GetAvatarURL()
			}
		} else {
			// External commenter
			if c.GetAuthorName() != nil {
				item.AuthorName = *c.GetAuthorName()
			}
			if c.GetAuthorEmail() != nil {
				item.AuthorEmail = *c.GetAuthorEmail()
			}
		}

		items = append(items, item)
	}

	return &GetCommentsResult{
		Items:  items,
		Total:  total,
		Limit:  q.Limit,
		Offset: q.Offset,
	}, nil
}
