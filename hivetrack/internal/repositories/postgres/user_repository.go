package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type UserRepository struct {
	ctx *DbContext
}

func NewUserRepository(ctx *DbContext) *UserRepository {
	return &UserRepository{ctx: ctx}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, sub, email, display_name, avatar_url, is_admin, created_at, last_login_at, xmin
		FROM users WHERE id=$1`, id)
	return scanUser(row)
}

func (r *UserRepository) GetBySub(ctx context.Context, sub string) (*models.User, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, sub, email, display_name, avatar_url, is_admin, created_at, last_login_at, xmin
		FROM users WHERE sub=$1`, sub)
	return scanUser(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, sub, email, display_name, avatar_url, is_admin, created_at, last_login_at, xmin
		FROM users WHERE email=$1`, email)
	return scanUser(row)
}

func (r *UserRepository) Upsert(ctx context.Context, user *models.User) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO users (id, sub, email, display_name, avatar_url, is_admin, created_at, last_login_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (sub) DO UPDATE SET
				email = EXCLUDED.email,
				display_name = EXCLUDED.display_name,
				avatar_url = EXCLUDED.avatar_url,
				last_login_at = EXCLUDED.last_login_at`,
			user.GetId(), user.GetSub(), user.GetEmail(), user.GetDisplayName(), user.GetAvatarURL(),
			user.GetIsAdmin(), user.GetCreatedAt(), user.GetLastLoginAt(),
		)
		if err != nil {
			return fmt.Errorf("upserting user: %w", err)
		}
		return nil
	})
}

func (r *UserRepository) List(ctx context.Context) ([]*models.User, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, sub, email, display_name, avatar_url, is_admin, created_at, last_login_at, xmin FROM users ORDER BY email`)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var id uuid.UUID
		var sub, email, displayName string
		var avatarURL sql.NullString
		var isAdmin bool
		var createdAt, lastLoginAt time.Time
		var xmin uint32
		if err := rows.Scan(&id, &sub, &email, &displayName, &avatarURL, &isAdmin, &createdAt, &lastLoginAt, &xmin); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		var avatarURLPtr *string
		if avatarURL.Valid {
			avatarURLPtr = &avatarURL.String
		}
		users = append(users, models.NewUserFromDB(id, createdAt, xmin, sub, email, displayName, avatarURLPtr, isAdmin, lastLoginAt))
	}
	return users, rows.Err()
}

func scanUser(row *sql.Row) (*models.User, error) {
	var id uuid.UUID
	var sub, email, displayName string
	var avatarURL sql.NullString
	var isAdmin bool
	var createdAt, lastLoginAt time.Time
	var xmin uint32

	err := row.Scan(&id, &sub, &email, &displayName, &avatarURL, &isAdmin, &createdAt, &lastLoginAt, &xmin)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning user: %w", err)
	}

	var avatarURLPtr *string
	if avatarURL.Valid {
		avatarURLPtr = &avatarURL.String
	}

	return models.NewUserFromDB(id, createdAt, xmin, sub, email, displayName, avatarURLPtr, isAdmin, lastLoginAt), nil
}
