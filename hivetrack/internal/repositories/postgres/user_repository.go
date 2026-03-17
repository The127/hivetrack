package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/the127/hivetrack/internal/models"
)

type UserRepository struct {
	ctx *DbContext
}

func NewUserRepository(ctx *DbContext) *UserRepository {
	return &UserRepository{ctx: ctx}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	s := sqlbuilder.Select("id", "sub", "email", "display_name", "avatar_url", "is_admin", "created_at", "last_login_at").
		From("users").
		Where(fmt.Sprintf("id = '%s'", id))

	query, args := s.BuildWithFlavor(sqlbuilder.PostgreSQL)
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx, query, args...)
	return scanUser(row)
}

func (r *UserRepository) GetBySub(ctx context.Context, sub string) (*models.User, error) {
	s := sqlbuilder.NewSelectBuilder().Select("id", "sub", "email", "display_name", "avatar_url", "is_admin", "created_at", "last_login_at").
		From("users")
	s.Where(s.Equal("sub", sub))

	query, args := s.BuildWithFlavor(sqlbuilder.PostgreSQL)
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx, query, args...)
	return scanUser(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	s := sqlbuilder.NewSelectBuilder().Select("id", "sub", "email", "display_name", "avatar_url", "is_admin", "created_at", "last_login_at").
		From("users")
	s.Where(s.Equal("email", email))

	query, args := s.BuildWithFlavor(sqlbuilder.PostgreSQL)
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx, query, args...)
	return scanUser(row)
}

func (r *UserRepository) Upsert(ctx context.Context, user *models.User) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users (id, sub, email, display_name, avatar_url, is_admin, created_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (sub) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url,
			last_login_at = EXCLUDED.last_login_at
	`

	_, err = tx.ExecContext(ctx, query,
		user.ID, user.Sub, user.Email, user.DisplayName, user.AvatarURL,
		user.IsAdmin, user.CreatedAt, user.LastLoginAt,
	)
	if err != nil {
		return fmt.Errorf("upserting user: %w", err)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context) ([]*models.User, error) {
	query := `SELECT id, sub, email, display_name, avatar_url, is_admin, created_at, last_login_at FROM users ORDER BY email`
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		var avatarURL sql.NullString
		var lastLogin time.Time
		if err := rows.Scan(&u.ID, &u.Sub, &u.Email, &u.DisplayName, &avatarURL, &u.IsAdmin, &u.CreatedAt, &lastLogin); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		if avatarURL.Valid {
			u.AvatarURL = &avatarURL.String
		}
		u.LastLoginAt = lastLogin
		users = append(users, &u)
	}
	return users, rows.Err()
}

func scanUser(row *sql.Row) (*models.User, error) {
	var u models.User
	var avatarURL sql.NullString
	err := row.Scan(&u.ID, &u.Sub, &u.Email, &u.DisplayName, &avatarURL, &u.IsAdmin, &u.CreatedAt, &u.LastLoginAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning user: %w", err)
	}
	if avatarURL.Valid {
		u.AvatarURL = &avatarURL.String
	}
	return &u, nil
}
