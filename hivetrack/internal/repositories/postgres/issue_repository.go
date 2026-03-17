package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type IssueRepository struct {
	ctx *DbContext
}

func NewIssueRepository(ctx *DbContext) *IssueRepository {
	return &IssueRepository{ctx: ctx}
}

func (r *IssueRepository) Insert(ctx context.Context, issue *models.Issue) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}

	checklistJSON, err := json.Marshal(issue.Checklist)
	if err != nil {
		return fmt.Errorf("marshaling checklist: %w", err)
	}

	q := `INSERT INTO issues (id, project_id, number, type, title, description, status,
		on_hold, hold_reason, hold_since, hold_note,
		priority, estimate, reporter_id, parent_id, milestone_id, sprint_id, sprint_carry_count,
		triaged, visibility, customer_email, customer_name, customer_token, checklist, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)`

	_, err = tx.ExecContext(ctx, q,
		issue.ID, issue.ProjectID, issue.Number, issue.Type, issue.Title, issue.Description, issue.Status,
		issue.OnHold, issue.HoldReason, issue.HoldSince, issue.HoldNote,
		issue.Priority, issue.Estimate, issue.ReporterID, issue.ParentID, issue.MilestoneID, issue.SprintID, issue.SprintCarryCount,
		issue.Triaged, issue.Visibility, issue.CustomerEmail, issue.CustomerName, issue.CustomerToken, checklistJSON,
		issue.CreatedAt, issue.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting issue: %w", err)
	}

	if err := r.syncAssignees(ctx, tx, issue.ID, issue.Assignees); err != nil {
		return err
	}
	if err := r.syncLabels(ctx, tx, issue.ID, issue.Labels); err != nil {
		return err
	}

	return nil
}

func (r *IssueRepository) Update(ctx context.Context, issue *models.Issue) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}

	checklistJSON, err := json.Marshal(issue.Checklist)
	if err != nil {
		return fmt.Errorf("marshaling checklist: %w", err)
	}

	q := `UPDATE issues SET title=$1, description=$2, status=$3,
		on_hold=$4, hold_reason=$5, hold_since=$6, hold_note=$7,
		priority=$8, estimate=$9, milestone_id=$10, sprint_id=$11, sprint_carry_count=$12,
		triaged=$13, visibility=$14, checklist=$15, updated_at=$16
		WHERE id=$17`

	res, err := tx.ExecContext(ctx, q,
		issue.Title, issue.Description, issue.Status,
		issue.OnHold, issue.HoldReason, issue.HoldSince, issue.HoldNote,
		issue.Priority, issue.Estimate, issue.MilestoneID, issue.SprintID, issue.SprintCarryCount,
		issue.Triaged, issue.Visibility, checklistJSON, issue.UpdatedAt,
		issue.ID,
	)
	if err != nil {
		return fmt.Errorf("updating issue: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("issue %s: %w", issue.ID, models.ErrNotFound)
	}

	if err := r.syncAssignees(ctx, tx, issue.ID, issue.Assignees); err != nil {
		return err
	}
	if err := r.syncLabels(ctx, tx, issue.ID, issue.Labels); err != nil {
		return err
	}

	return nil
}

func (r *IssueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM issues WHERE id=$1`, id)
	return err
}

func (r *IssueRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Issue, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx, issueSelectQuery+` WHERE i.id=$1`, id)
	return scanIssue(row)
}

func (r *IssueRepository) GetByNumber(ctx context.Context, projectID uuid.UUID, number int) (*models.Issue, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx, issueSelectQuery+` WHERE i.project_id=$1 AND i.number=$2`, projectID, number)
	return scanIssue(row)
}

func (r *IssueRepository) List(ctx context.Context, filter *repositories.IssueFilter) ([]*models.Issue, int, error) {
	baseQuery := issueSelectQuery + ` WHERE 1=1`
	var args []any
	argIdx := 1

	if filter.ProjectID != nil {
		baseQuery += fmt.Sprintf(` AND i.project_id=$%d`, argIdx)
		args = append(args, *filter.ProjectID)
		argIdx++
	}
	if filter.Status != nil {
		baseQuery += fmt.Sprintf(` AND i.status=$%d`, argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Priority != nil {
		baseQuery += fmt.Sprintf(` AND i.priority=$%d`, argIdx)
		args = append(args, *filter.Priority)
		argIdx++
	}
	if filter.SprintID != nil {
		baseQuery += fmt.Sprintf(` AND i.sprint_id=$%d`, argIdx)
		args = append(args, *filter.SprintID)
		argIdx++
	}
	if filter.Triaged != nil {
		baseQuery += fmt.Sprintf(` AND i.triaged=$%d`, argIdx)
		args = append(args, *filter.Triaged)
		argIdx++
	}
	if filter.AssigneeID != nil {
		baseQuery += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM issue_assignees ia WHERE ia.issue_id = i.id AND ia.user_id=$%d)`, argIdx)
		args = append(args, *filter.AssigneeID)
		argIdx++
	}
	if filter.Text != nil && *filter.Text != "" {
		baseQuery += fmt.Sprintf(` AND to_tsvector('english', i.title || ' ' || coalesce(i.description, '')) @@ plainto_tsquery('english', $%d)`, argIdx)
		args = append(args, *filter.Text)
		argIdx++
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM (` + baseQuery + `) sub`
	var total int
	if err := r.ctx.queryContext(ctx).QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting issues: %w", err)
	}

	baseQuery += ` ORDER BY i.created_at DESC`

	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(` LIMIT %d`, filter.Limit)
	}
	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(` OFFSET %d`, filter.Offset)
	}

	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing issues: %w", err)
	}
	defer rows.Close()

	var issues []*models.Issue
	for rows.Next() {
		issue, err := scanIssueRow(rows)
		if err != nil {
			return nil, 0, err
		}
		issues = append(issues, issue)
	}
	return issues, total, rows.Err()
}

func (r *IssueRepository) syncAssignees(ctx context.Context, tx *sql.Tx, issueID uuid.UUID, assignees []uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM issue_assignees WHERE issue_id=$1`, issueID); err != nil {
		return fmt.Errorf("clearing assignees: %w", err)
	}
	for _, uid := range assignees {
		if _, err := tx.ExecContext(ctx, `INSERT INTO issue_assignees (issue_id, user_id) VALUES ($1,$2)`, issueID, uid); err != nil {
			return fmt.Errorf("inserting assignee: %w", err)
		}
	}
	return nil
}

func (r *IssueRepository) syncLabels(ctx context.Context, tx *sql.Tx, issueID uuid.UUID, labels []uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM issue_labels WHERE issue_id=$1`, issueID); err != nil {
		return fmt.Errorf("clearing labels: %w", err)
	}
	for _, lid := range labels {
		if _, err := tx.ExecContext(ctx, `INSERT INTO issue_labels (issue_id, label_id) VALUES ($1,$2)`, issueID, lid); err != nil {
			return fmt.Errorf("inserting label: %w", err)
		}
	}
	return nil
}

const issueSelectQuery = `
	SELECT i.id, i.project_id, i.number, i.type, i.title, i.description, i.status,
		i.on_hold, i.hold_reason, i.hold_since, i.hold_note,
		i.priority, i.estimate, i.reporter_id, i.parent_id, i.milestone_id, i.sprint_id, i.sprint_carry_count,
		i.triaged, i.visibility, i.customer_email, i.customer_name, i.customer_token,
		i.checklist, i.created_at, i.updated_at,
		COALESCE((SELECT array_agg(user_id) FROM issue_assignees WHERE issue_id=i.id), '{}'),
		COALESCE((SELECT array_agg(label_id) FROM issue_labels WHERE issue_id=i.id), '{}')
	FROM issues i`

func scanIssue(row *sql.Row) (*models.Issue, error) {
	var i models.Issue
	var desc, holdReason, holdNote, customerEmail, customerName sql.NullString
	var holdSince sql.NullTime
	var customerToken uuid.NullUUID
	var reporterID, parentID, milestoneID, sprintID uuid.NullUUID
	var checklistJSON []byte
	var assigneeArr, labelArr []byte // will parse as postgres array

	err := row.Scan(
		&i.ID, &i.ProjectID, &i.Number, &i.Type, &i.Title, &desc, &i.Status,
		&i.OnHold, &holdReason, &holdSince, &holdNote,
		&i.Priority, &i.Estimate, &reporterID, &parentID, &milestoneID, &sprintID, &i.SprintCarryCount,
		&i.Triaged, &i.Visibility, &customerEmail, &customerName, &customerToken,
		&checklistJSON, &i.CreatedAt, &i.UpdatedAt,
		&assigneeArr, &labelArr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning issue: %w", err)
	}

	if desc.Valid {
		i.Description = &desc.String
	}
	if holdReason.Valid {
		hr := models.HoldReason(holdReason.String)
		i.HoldReason = &hr
	}
	if holdSince.Valid {
		i.HoldSince = &holdSince.Time
	}
	if holdNote.Valid {
		i.HoldNote = &holdNote.String
	}
	if reporterID.Valid {
		i.ReporterID = &reporterID.UUID
	}
	if parentID.Valid {
		i.ParentID = &parentID.UUID
	}
	if milestoneID.Valid {
		i.MilestoneID = &milestoneID.UUID
	}
	if sprintID.Valid {
		i.SprintID = &sprintID.UUID
	}
	if customerToken.Valid {
		i.CustomerToken = &customerToken.UUID
	}
	if customerEmail.Valid {
		i.CustomerEmail = &customerEmail.String
	}
	if customerName.Valid {
		i.CustomerName = &customerName.String
	}

	if err := json.Unmarshal(checklistJSON, &i.Checklist); err != nil {
		i.Checklist = []models.ChecklistItem{}
	}

	// Parse UUID arrays from postgres format
	i.Assignees = parseUUIDArray(assigneeArr)
	i.Labels = parseUUIDArray(labelArr)

	return &i, nil
}

func scanIssueRow(rows *sql.Rows) (*models.Issue, error) {
	var i models.Issue
	var desc, holdReason, holdNote, customerEmail, customerName sql.NullString
	var holdSince sql.NullTime
	var customerToken uuid.NullUUID
	var reporterID, parentID, milestoneID, sprintID uuid.NullUUID
	var checklistJSON []byte
	var assigneeArr, labelArr []byte

	err := rows.Scan(
		&i.ID, &i.ProjectID, &i.Number, &i.Type, &i.Title, &desc, &i.Status,
		&i.OnHold, &holdReason, &holdSince, &holdNote,
		&i.Priority, &i.Estimate, &reporterID, &parentID, &milestoneID, &sprintID, &i.SprintCarryCount,
		&i.Triaged, &i.Visibility, &customerEmail, &customerName, &customerToken,
		&checklistJSON, &i.CreatedAt, &i.UpdatedAt,
		&assigneeArr, &labelArr,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning issue row: %w", err)
	}

	if desc.Valid {
		i.Description = &desc.String
	}
	if holdReason.Valid {
		hr := models.HoldReason(holdReason.String)
		i.HoldReason = &hr
	}
	if holdSince.Valid {
		i.HoldSince = &holdSince.Time
	}
	if holdNote.Valid {
		i.HoldNote = &holdNote.String
	}
	if reporterID.Valid {
		i.ReporterID = &reporterID.UUID
	}
	if parentID.Valid {
		i.ParentID = &parentID.UUID
	}
	if milestoneID.Valid {
		i.MilestoneID = &milestoneID.UUID
	}
	if sprintID.Valid {
		i.SprintID = &sprintID.UUID
	}
	if customerToken.Valid {
		i.CustomerToken = &customerToken.UUID
	}
	if customerEmail.Valid {
		i.CustomerEmail = &customerEmail.String
	}
	if customerName.Valid {
		i.CustomerName = &customerName.String
	}

	if err := json.Unmarshal(checklistJSON, &i.Checklist); err != nil {
		i.Checklist = []models.ChecklistItem{}
	}

	i.Assignees = parseUUIDArray(assigneeArr)
	i.Labels = parseUUIDArray(labelArr)

	return &i, nil
}

// parseUUIDArray parses postgres UUID array format like {uuid1,uuid2} or {} into []uuid.UUID.
func parseUUIDArray(data []byte) []uuid.UUID {
	if len(data) == 0 {
		return nil
	}
	s := string(data)
	if s == "{}" {
		return nil
	}
	// Remove braces
	s = s[1 : len(s)-1]
	parts := splitCSV(s)
	result := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		id, err := uuid.Parse(p)
		if err == nil {
			result = append(result, id)
		}
	}
	return result
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	var current []byte
	for _, c := range s {
		if c == ',' {
			result = append(result, string(current))
			current = current[:0]
		} else {
			current = append(current, byte(c))
		}
	}
	result = append(result, string(current))
	return result
}
