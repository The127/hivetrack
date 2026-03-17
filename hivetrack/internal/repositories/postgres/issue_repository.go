package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type IssueRepository struct {
	ctx *DbContext
}

func NewIssueRepository(ctx *DbContext) *IssueRepository {
	return &IssueRepository{ctx: ctx}
}

func (r *IssueRepository) Insert(issue *models.Issue) {
	r.ctx.changeTracker.Add(change.NewEntry(issueEntityType, issue, change.Added))
}

func (r *IssueRepository) Update(issue *models.Issue) {
	r.ctx.changeTracker.Add(change.NewEntry(issueEntityType, issue, change.Updated))
}

func (r *IssueRepository) Delete(issue *models.Issue) {
	r.ctx.changeTracker.Add(change.NewEntry(issueEntityType, issue, change.Deleted))
}

func (r *IssueRepository) ExecuteInsert(ctx context.Context, tx *sql.Tx, issue *models.Issue) error {
	checklistJSON, err := json.Marshal(issue.GetChecklist())
	if err != nil {
		return fmt.Errorf("marshaling checklist: %w", err)
	}

	var version int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO issues (id, project_id, number, type, title, description, status,
		on_hold, hold_reason, hold_since, hold_note,
		priority, estimate, reporter_id, parent_id, milestone_id, sprint_id, sprint_carry_count,
		triaged, visibility, customer_email, customer_name, customer_token, "rank", checklist, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)
		RETURNING version`,
		issue.GetId(), issue.GetProjectID(), issue.GetNumber(), issue.GetType(),
		issue.GetTitle(), issue.GetDescription(), issue.GetStatus(),
		issue.GetOnHold(), issue.GetHoldReason(), issue.GetHoldSince(), issue.GetHoldNote(),
		issue.GetPriority(), issue.GetEstimate(), issue.GetReporterID(), issue.GetParentID(),
		issue.GetMilestoneID(), issue.GetSprintID(), issue.GetSprintCarryCount(),
		issue.GetTriaged(), issue.GetVisibility(),
		issue.GetCustomerEmail(), issue.GetCustomerName(), issue.GetCustomerToken(),
		issue.GetRank(), checklistJSON, issue.GetCreatedAt(), issue.GetUpdatedAt(),
	).Scan(&version)
	if err != nil {
		return fmt.Errorf("inserting issue: %w", err)
	}

	if err := r.syncAssignees(ctx, tx, issue.GetId(), issue.GetAssignees()); err != nil {
		return err
	}
	if err := r.syncLabels(ctx, tx, issue.GetId(), issue.GetLabels()); err != nil {
		return err
	}

	issue.SetVersion(version)
	issue.ClearChanges()
	return nil
}

func (r *IssueRepository) ExecuteUpdate(ctx context.Context, tx *sql.Tx, issue *models.Issue) error {
	if !issue.HasChanges() {
		return nil
	}

	// Capture junction table needs before building query
	doSyncAssignees := issue.HasChange(models.IssueChangeAssignees)
	doSyncLabels := issue.HasChange(models.IssueChangeLabels)

	var setClauses []string
	var args []any
	argIdx := 1

	// updated_at is always included on update
	setClauses = append(setClauses, fmt.Sprintf("updated_at=$%d", argIdx))
	args = append(args, issue.GetUpdatedAt())
	argIdx++

	if issue.HasChange(models.IssueChangeTitle) {
		setClauses = append(setClauses, fmt.Sprintf("title=$%d", argIdx))
		args = append(args, issue.GetTitle())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeDescription) {
		setClauses = append(setClauses, fmt.Sprintf("description=$%d", argIdx))
		args = append(args, issue.GetDescription())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeStatus) {
		setClauses = append(setClauses, fmt.Sprintf("status=$%d", argIdx))
		args = append(args, issue.GetStatus())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeHold) {
		setClauses = append(setClauses, fmt.Sprintf("on_hold=$%d", argIdx))
		args = append(args, issue.GetOnHold())
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("hold_reason=$%d", argIdx))
		args = append(args, issue.GetHoldReason())
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("hold_since=$%d", argIdx))
		args = append(args, issue.GetHoldSince())
		argIdx++
		setClauses = append(setClauses, fmt.Sprintf("hold_note=$%d", argIdx))
		args = append(args, issue.GetHoldNote())
		argIdx++
	}
	if issue.HasChange(models.IssueChangePriority) {
		setClauses = append(setClauses, fmt.Sprintf("priority=$%d", argIdx))
		args = append(args, issue.GetPriority())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeEstimate) {
		setClauses = append(setClauses, fmt.Sprintf("estimate=$%d", argIdx))
		args = append(args, issue.GetEstimate())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeMilestoneID) {
		setClauses = append(setClauses, fmt.Sprintf("milestone_id=$%d", argIdx))
		args = append(args, issue.GetMilestoneID())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeSprintID) {
		setClauses = append(setClauses, fmt.Sprintf("sprint_id=$%d", argIdx))
		args = append(args, issue.GetSprintID())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeSprintCarryCount) {
		setClauses = append(setClauses, fmt.Sprintf("sprint_carry_count=$%d", argIdx))
		args = append(args, issue.GetSprintCarryCount())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeTriaged) {
		setClauses = append(setClauses, fmt.Sprintf("triaged=$%d", argIdx))
		args = append(args, issue.GetTriaged())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeVisibility) {
		setClauses = append(setClauses, fmt.Sprintf("visibility=$%d", argIdx))
		args = append(args, issue.GetVisibility())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeChecklist) {
		checklistJSON, err := json.Marshal(issue.GetChecklist())
		if err != nil {
			return fmt.Errorf("marshaling checklist: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("checklist=$%d", argIdx))
		args = append(args, checklistJSON)
		argIdx++
	}
	if issue.HasChange(models.IssueChangeRank) {
		setClauses = append(setClauses, fmt.Sprintf(`"rank"=$%d`, argIdx))
		args = append(args, issue.GetRank())
		argIdx++
	}
	if issue.HasChange(models.IssueChangeParentID) {
		setClauses = append(setClauses, fmt.Sprintf("parent_id=$%d", argIdx))
		args = append(args, issue.GetParentID())
		argIdx++
	}

	setClauses = append(setClauses, "version = version + 1")

	query := fmt.Sprintf("UPDATE issues SET %s WHERE id=$%d", strings.Join(setClauses, ", "), argIdx)
	args = append(args, issue.GetId())
	argIdx++

	if issue.GetVersion() != nil {
		query += fmt.Sprintf(" AND version=$%d", argIdx)
		args = append(args, issue.GetVersion().(int))
		argIdx++
	}
	query += " RETURNING version"

	var version int
	err := tx.QueryRowContext(ctx, query, args...).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		if issue.GetVersion() != nil {
			return fmt.Errorf("issue %s: %w", issue.GetId(), models.ErrConcurrentUpdate)
		}
		return fmt.Errorf("issue %s: %w", issue.GetId(), models.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("updating issue: %w", err)
	}

	if doSyncAssignees {
		if err := r.syncAssignees(ctx, tx, issue.GetId(), issue.GetAssignees()); err != nil {
			return err
		}
	}
	if doSyncLabels {
		if err := r.syncLabels(ctx, tx, issue.GetId(), issue.GetLabels()); err != nil {
			return err
		}
	}

	issue.SetVersion(version)
	issue.ClearChanges()
	return nil
}

func (r *IssueRepository) ExecuteDelete(ctx context.Context, tx *sql.Tx, issue *models.Issue) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM issues WHERE id=$1`, issue.GetId())
	if err != nil {
		return fmt.Errorf("deleting issue: %w", err)
	}
	return nil
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
	if filter.InBacklog != nil && *filter.InBacklog {
		baseQuery += ` AND i.sprint_id IS NULL`
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
	if filter.Type != nil {
		baseQuery += fmt.Sprintf(` AND i.type=$%d`, argIdx)
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.ParentID != nil {
		baseQuery += fmt.Sprintf(` AND i.parent_id=$%d`, argIdx)
		args = append(args, *filter.ParentID)
		argIdx++
	}
	if filter.HasNoParent != nil && *filter.HasNoParent {
		baseQuery += ` AND i.parent_id IS NULL`
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM (` + baseQuery + `) sub`
	var total int
	if err := r.ctx.queryContext(ctx).QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting issues: %w", err)
	}

	baseQuery += ` ORDER BY i."rank" ASC NULLS LAST, i.created_at DESC`

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
		i."rank", i.checklist, i.created_at, i.updated_at, i.version,
		COALESCE((SELECT array_agg(user_id) FROM issue_assignees WHERE issue_id=i.id), '{}'),
		COALESCE((SELECT array_agg(label_id) FROM issue_labels WHERE issue_id=i.id), '{}')
	FROM issues i`

func scanIssue(row *sql.Row) (*models.Issue, error) {
	var id, projectID uuid.UUID
	var number int
	var issueType models.IssueType
	var title string
	var desc, holdReason, holdNote, customerEmail, customerName, rankStr sql.NullString
	var status models.IssueStatus
	var onHold bool
	var holdSince sql.NullTime
	var priority models.IssuePriority
	var estimate models.IssueEstimate
	var reporterID, parentID, milestoneID, sprintID uuid.NullUUID
	var customerToken uuid.NullUUID
	var sprintCarryCount int
	var triaged bool
	var visibility models.IssueVisibility
	var checklistJSON []byte
	var createdAt, updatedAt time.Time
	var version int
	var assigneeArr, labelArr []byte

	err := row.Scan(
		&id, &projectID, &number, &issueType, &title, &desc, &status,
		&onHold, &holdReason, &holdSince, &holdNote,
		&priority, &estimate, &reporterID, &parentID, &milestoneID, &sprintID, &sprintCarryCount,
		&triaged, &visibility, &customerEmail, &customerName, &customerToken,
		&rankStr, &checklistJSON, &createdAt, &updatedAt, &version,
		&assigneeArr, &labelArr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning issue: %w", err)
	}

	return buildIssue(
		id, projectID, number, issueType, title,
		desc, holdReason, holdNote, customerEmail, customerName,
		status, onHold, holdSince, customerToken,
		priority, estimate,
		reporterID, parentID, milestoneID, sprintID,
		sprintCarryCount, triaged, visibility,
		rankStr, checklistJSON, createdAt, updatedAt, version,
		assigneeArr, labelArr,
	), nil
}

func scanIssueRow(rows *sql.Rows) (*models.Issue, error) {
	var id, projectID uuid.UUID
	var number int
	var issueType models.IssueType
	var title string
	var desc, holdReason, holdNote, customerEmail, customerName, rankStr sql.NullString
	var status models.IssueStatus
	var onHold bool
	var holdSince sql.NullTime
	var priority models.IssuePriority
	var estimate models.IssueEstimate
	var reporterID, parentID, milestoneID, sprintID uuid.NullUUID
	var customerToken uuid.NullUUID
	var sprintCarryCount int
	var triaged bool
	var visibility models.IssueVisibility
	var checklistJSON []byte
	var createdAt, updatedAt time.Time
	var version int
	var assigneeArr, labelArr []byte

	err := rows.Scan(
		&id, &projectID, &number, &issueType, &title, &desc, &status,
		&onHold, &holdReason, &holdSince, &holdNote,
		&priority, &estimate, &reporterID, &parentID, &milestoneID, &sprintID, &sprintCarryCount,
		&triaged, &visibility, &customerEmail, &customerName, &customerToken,
		&rankStr, &checklistJSON, &createdAt, &updatedAt, &version,
		&assigneeArr, &labelArr,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning issue row: %w", err)
	}

	return buildIssue(
		id, projectID, number, issueType, title,
		desc, holdReason, holdNote, customerEmail, customerName,
		status, onHold, holdSince, customerToken,
		priority, estimate,
		reporterID, parentID, milestoneID, sprintID,
		sprintCarryCount, triaged, visibility,
		rankStr, checklistJSON, createdAt, updatedAt, version,
		assigneeArr, labelArr,
	), nil
}

func buildIssue(
	id, projectID uuid.UUID, number int, issueType models.IssueType, title string,
	desc, holdReason, holdNote, customerEmail, customerName sql.NullString,
	status models.IssueStatus, onHold bool, holdSince sql.NullTime,
	customerToken uuid.NullUUID,
	priority models.IssuePriority, estimate models.IssueEstimate,
	reporterID, parentID, milestoneID, sprintID uuid.NullUUID,
	sprintCarryCount int, triaged bool, visibility models.IssueVisibility,
	rankStr sql.NullString,
	checklistJSON []byte, createdAt, updatedAt time.Time, version int,
	assigneeArr, labelArr []byte,
) *models.Issue {
	var descPtr *string
	if desc.Valid {
		descPtr = &desc.String
	}
	var holdReasonPtr *models.HoldReason
	if holdReason.Valid {
		hr := models.HoldReason(holdReason.String)
		holdReasonPtr = &hr
	}
	var holdSincePtr *time.Time
	if holdSince.Valid {
		holdSincePtr = &holdSince.Time
	}
	var holdNotePtr *string
	if holdNote.Valid {
		holdNotePtr = &holdNote.String
	}
	var reporterIDPtr *uuid.UUID
	if reporterID.Valid {
		reporterIDPtr = &reporterID.UUID
	}
	var parentIDPtr *uuid.UUID
	if parentID.Valid {
		parentIDPtr = &parentID.UUID
	}
	var milestoneIDPtr *uuid.UUID
	if milestoneID.Valid {
		milestoneIDPtr = &milestoneID.UUID
	}
	var sprintIDPtr *uuid.UUID
	if sprintID.Valid {
		sprintIDPtr = &sprintID.UUID
	}
	var customerTokenPtr *uuid.UUID
	if customerToken.Valid {
		customerTokenPtr = &customerToken.UUID
	}
	var customerEmailPtr *string
	if customerEmail.Valid {
		customerEmailPtr = &customerEmail.String
	}
	var customerNamePtr *string
	if customerName.Valid {
		customerNamePtr = &customerName.String
	}
	var rankPtr *string
	if rankStr.Valid {
		rankPtr = &rankStr.String
	}

	var checklist []models.ChecklistItem
	if err := json.Unmarshal(checklistJSON, &checklist); err != nil {
		checklist = []models.ChecklistItem{}
	}

	assignees := parseUUIDArray(assigneeArr)
	labels := parseUUIDArray(labelArr)

	return models.NewIssueFromDB(
		id, createdAt, updatedAt, version,
		projectID, number, issueType, title, descPtr, status,
		onHold, holdReasonPtr, holdSincePtr, holdNotePtr,
		priority, estimate,
		reporterIDPtr, parentIDPtr, milestoneIDPtr, sprintIDPtr,
		sprintCarryCount, triaged, visibility,
		customerEmailPtr, customerNamePtr, customerTokenPtr,
		rankPtr, checklist, assignees, labels, nil,
	)
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
