package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/the127/hivetrack/internal/repositories"
)

type IssueStatusLogRepository struct {
	ctx *DbContext
}

func NewIssueStatusLogRepository(ctx *DbContext) *IssueStatusLogRepository {
	return &IssueStatusLogRepository{ctx: ctx}
}

func (r *IssueStatusLogRepository) Insert(ctx context.Context, issueID uuid.UUID, status string, changedAt time.Time) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		id := uuid.New()
		_, err := tx.ExecContext(ctx,
			`INSERT INTO issue_status_log (id, issue_id, status, changed_at) VALUES ($1, $2, $3, $4)`,
			id, issueID, status, changedAt,
		)
		return err
	})
}

func (r *IssueStatusLogRepository) GetBurndownPoints(
	ctx context.Context,
	sprintID uuid.UUID,
	startDate, endDate time.Time,
	terminalStatuses []string,
) ([]repositories.BurndownPoint, error) {
	// Cap end date at today
	today := time.Now().UTC().Truncate(24 * time.Hour)
	effectiveEnd := endDate
	if effectiveEnd.After(today) {
		effectiveEnd = today
	}

	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx, `
		SELECT
		    gs.day::date AS date,
		    COUNT(i.id) FILTER (
		        WHERE NOT EXISTS (
		            SELECT 1 FROM issue_status_log sl
		            WHERE sl.issue_id = i.id
		              AND sl.status = ANY($2)
		              AND sl.changed_at::date <= gs.day::date
		        )
		    ) AS remaining
		FROM generate_series($3::timestamp, $4::timestamp, '1 day') AS gs(day)
		CROSS JOIN (SELECT id FROM issues WHERE sprint_id = $1 AND type = 'task') i
		GROUP BY gs.day
		ORDER BY gs.day`,
		sprintID, pq.Array(terminalStatuses), startDate, effectiveEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("getting burndown points: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var points []repositories.BurndownPoint
	for rows.Next() {
		var p repositories.BurndownPoint
		if err := rows.Scan(&p.Date, &p.Remaining); err != nil {
			return nil, fmt.Errorf("scanning burndown point: %w", err)
		}
		points = append(points, p)
	}
	return points, rows.Err()
}
