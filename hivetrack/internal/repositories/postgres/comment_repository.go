package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
)

type CommentRepository struct {
	ctx *DbContext
}

func NewCommentRepository(ctx *DbContext) *CommentRepository {
	return &CommentRepository{ctx: ctx}
}

func (r *CommentRepository) Insert(comment *models.Comment) {
	r.ctx.changeTracker.Add(change.NewEntry(commentEntityType, comment, change.Added))
}

func (r *CommentRepository) Update(comment *models.Comment) {
	r.ctx.changeTracker.Add(change.NewEntry(commentEntityType, comment, change.Updated))
}

func (r *CommentRepository) Delete(comment *models.Comment) {
	r.ctx.changeTracker.Add(change.NewEntry(commentEntityType, comment, change.Deleted))
}

func (r *CommentRepository) ExecuteInsert(ctx context.Context, tx *sql.Tx, comment *models.Comment) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO comments (id, issue_id, author_id, author_email, author_name, body, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		comment.GetId(), comment.GetIssueID(), comment.GetAuthorID(),
		comment.GetAuthorEmail(), comment.GetAuthorName(), comment.GetBody(),
		comment.GetCreatedAt(), comment.GetUpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("inserting comment: %w", err)
	}
	comment.ClearChanges()
	return nil
}

func (r *CommentRepository) ExecuteUpdate(ctx context.Context, tx *sql.Tx, comment *models.Comment) error {
	if !comment.HasChanges() {
		return nil
	}

	var setClauses []string
	var args []any
	argIdx := 1

	if comment.HasChange(models.CommentChangeBody) {
		setClauses = append(setClauses, fmt.Sprintf("body=$%d", argIdx))
		args = append(args, comment.GetBody())
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at=$%d", argIdx))
	args = append(args, comment.GetUpdatedAt())
	argIdx++

	query := fmt.Sprintf("UPDATE comments SET %s WHERE id=$%d", strings.Join(setClauses, ", "), argIdx)
	args = append(args, comment.GetId())

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating comment: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("comment %s: %w", comment.GetId(), models.ErrNotFound)
	}

	comment.ClearChanges()
	return nil
}

func (r *CommentRepository) ExecuteDelete(ctx context.Context, tx *sql.Tx, comment *models.Comment) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM comments WHERE id=$1`, comment.GetId())
	if err != nil {
		return fmt.Errorf("deleting comment: %w", err)
	}
	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, issue_id, author_id, author_email, author_name, body, created_at, updated_at FROM comments WHERE id=$1`, id)
	return scanComment(row)
}

func (r *CommentRepository) List(ctx context.Context, issueID uuid.UUID, limit, offset int) ([]*models.Comment, int, error) {
	// Count total
	var total int
	err := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM comments WHERE issue_id=$1`, issueID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting comments: %w", err)
	}

	query := `SELECT id, issue_id, author_id, author_email, author_name, body, created_at, updated_at FROM comments WHERE issue_id=$1 ORDER BY created_at ASC`
	var args []any
	args = append(args, issueID)
	argIdx := 2

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
		argIdx++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, offset)
	}

	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing comments: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var id, issueID uuid.UUID
		var authorID *uuid.UUID
		var authorEmail, authorName *string
		var body string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &issueID, &authorID, &authorEmail, &authorName, &body, &createdAt, &updatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning comment: %w", err)
		}
		comments = append(comments, models.NewCommentFromDB(id, createdAt, updatedAt, issueID, authorID, authorEmail, authorName, body))
	}
	return comments, total, rows.Err()
}

func scanComment(row *sql.Row) (*models.Comment, error) {
	var id, issueID uuid.UUID
	var authorID *uuid.UUID
	var authorEmail, authorName *string
	var body string
	var createdAt, updatedAt time.Time

	err := row.Scan(&id, &issueID, &authorID, &authorEmail, &authorName, &body, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning comment: %w", err)
	}

	return models.NewCommentFromDB(id, createdAt, updatedAt, issueID, authorID, authorEmail, authorName, body), nil
}
