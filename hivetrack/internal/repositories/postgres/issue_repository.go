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
		priority, estimate, reporter_id, owner_id, parent_id, milestone_id, sprint_id, sprint_carry_count,
		triaged, refined, visibility, customer_email, customer_name, customer_token, "rank", cancel_reason, checklist, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)
		RETURNING version`,
		issue.GetId(), issue.GetProjectID(), issue.GetNumber(), issue.GetType(),
		issue.GetTitle(), issue.GetDescription(), issue.GetStatus(),
		issue.GetOnHold(), issue.GetHoldReason(), issue.GetHoldSince(), issue.GetHoldNote(),
		issue.GetPriority(), issue.GetEstimate(), issue.GetReporterID(), issue.GetOwnerID(), issue.GetParentID(),
		issue.GetMilestoneID(), issue.GetSprintID(), issue.GetSprintCarryCount(),
		issue.GetTriaged(), issue.GetRefined(), issue.GetVisibility(),
		issue.GetCustomerEmail(), issue.GetCustomerName(), issue.GetCustomerToken(),
		issue.GetRank(), issue.GetCancelReason(), checklistJSON, issue.GetCreatedAt(), issue.GetUpdatedAt(),
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

// issueColumnMapping defines the change-to-column mapping for simple single-column fields.
// Each entry maps a change constant to a column name and a function that extracts the value.
// To add a new single-column field, append one entry here.
var issueColumnMapping = []struct {
	change models.IssueChange
	column string
	value  func(*models.Issue) any
}{
	{models.IssueChangeTitle, "title", func(i *models.Issue) any { return i.GetTitle() }},
	{models.IssueChangeDescription, "description", func(i *models.Issue) any { return i.GetDescription() }},
	{models.IssueChangeStatus, "status", func(i *models.Issue) any { return i.GetStatus() }},
	{models.IssueChangePriority, "priority", func(i *models.Issue) any { return i.GetPriority() }},
	{models.IssueChangeEstimate, "estimate", func(i *models.Issue) any { return i.GetEstimate() }},
	{models.IssueChangeMilestoneID, "milestone_id", func(i *models.Issue) any { return i.GetMilestoneID() }},
	{models.IssueChangeSprintID, "sprint_id", func(i *models.Issue) any { return i.GetSprintID() }},
	{models.IssueChangeSprintCarryCount, "sprint_carry_count", func(i *models.Issue) any { return i.GetSprintCarryCount() }},
	{models.IssueChangeTriaged, "triaged", func(i *models.Issue) any { return i.GetTriaged() }},
	{models.IssueChangeRefined, "refined", func(i *models.Issue) any { return i.GetRefined() }},
	{models.IssueChangeVisibility, "visibility", func(i *models.Issue) any { return i.GetVisibility() }},
	{models.IssueChangeRank, `"rank"`, func(i *models.Issue) any { return i.GetRank() }},
	{models.IssueChangeParentID, "parent_id", func(i *models.Issue) any { return i.GetParentID() }},
	{models.IssueChangeOwnerID, "owner_id", func(i *models.Issue) any { return i.GetOwnerID() }},
	{models.IssueChangeCancelReason, "cancel_reason", func(i *models.Issue) any { return i.GetCancelReason() }},
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

	// Simple single-column fields driven by the mapping table
	for _, m := range issueColumnMapping {
		if issue.HasChange(m.change) {
			setClauses = append(setClauses, fmt.Sprintf("%s=$%d", m.column, argIdx))
			args = append(args, m.value(issue))
			argIdx++
		}
	}

	// Hold maps one change to four columns
	if issue.HasChange(models.IssueChangeHold) {
		for _, col := range []struct {
			name string
			val  any
		}{
			{"on_hold", issue.GetOnHold()},
			{"hold_reason", issue.GetHoldReason()},
			{"hold_since", issue.GetHoldSince()},
			{"hold_note", issue.GetHoldNote()},
		} {
			setClauses = append(setClauses, fmt.Sprintf("%s=$%d", col.name, argIdx))
			args = append(args, col.val)
			argIdx++
		}
	}

	// Checklist requires JSON marshaling
	if issue.HasChange(models.IssueChangeChecklist) {
		checklistJSON, err := json.Marshal(issue.GetChecklist())
		if err != nil {
			return fmt.Errorf("marshaling checklist: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("checklist=$%d", argIdx))
		args = append(args, checklistJSON)
		argIdx++
	}

	setClauses = append(setClauses, "version = version + 1")

	query := fmt.Sprintf("UPDATE issues SET %s WHERE id=$%d", strings.Join(setClauses, ", "), argIdx) //nolint:gosec
	args = append(args, issue.GetId())
	argIdx++

	if issue.GetVersion() != nil {
		query += fmt.Sprintf(" AND version=$%d", argIdx)
		args = append(args, issue.GetVersion().(int))
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

func (r *IssueRepository) InsertLink(ctx context.Context, link models.IssueLink) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO issue_links (id, source_issue_id, target_issue_id, link_type) VALUES ($1,$2,$3,$4)`,
			link.ID, link.SourceIssueID, link.TargetIssueID, link.LinkType,
		)
		if err != nil {
			return fmt.Errorf("inserting issue link: %w", err)
		}
		return nil
	})
}

func (r *IssueRepository) ListLinks(ctx context.Context, issueID uuid.UUID) ([]models.IssueLink, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, source_issue_id, target_issue_id, link_type FROM issue_links WHERE source_issue_id=$1 OR target_issue_id=$1`,
		issueID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing issue links: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var links []models.IssueLink
	for rows.Next() {
		var l models.IssueLink
		if err := rows.Scan(&l.ID, &l.SourceIssueID, &l.TargetIssueID, &l.LinkType); err != nil {
			return nil, fmt.Errorf("scanning issue link: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (r *IssueRepository) CountUntriagedByProject(ctx context.Context) (map[uuid.UUID]int, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT project_id, COUNT(*) FROM issues WHERE triaged = false GROUP BY project_id`)
	if err != nil {
		return nil, fmt.Errorf("counting untriaged issues: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[uuid.UUID]int)
	for rows.Next() {
		var projectID uuid.UUID
		var count int
		if err := rows.Scan(&projectID, &count); err != nil {
			return nil, fmt.Errorf("scanning untriaged count: %w", err)
		}
		result[projectID] = count
	}
	return result, rows.Err()
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
	w := newWhereBuilder(issueSelectQuery)

	r.applyIssueFilters(w, filter)

	// Count total
	countQuery := `SELECT COUNT(*) FROM (` + w.query + `) sub`
	var total int
	if err := r.ctx.queryContext(ctx).QueryRowContext(ctx, countQuery, w.args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting issues: %w", err)
	}

	w.query += ` ORDER BY i."rank" ASC NULLS LAST, i.created_at DESC`

	if filter.Limit > 0 {
		w.query += fmt.Sprintf(` LIMIT %d`, filter.Limit)
	}
	if filter.Offset > 0 {
		w.query += fmt.Sprintf(` OFFSET %d`, filter.Offset)
	}

	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx, w.query, w.args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing issues: %w", err)
	}
	defer func() { _ = rows.Close() }()

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

func (r *IssueRepository) applyIssueFilters(w *whereBuilder, filter *repositories.IssueFilter) {
	w.eq(`i.project_id`, filter.ProjectID != nil, filter.ProjectID)
	w.eq(`i.status`, filter.Status != nil, filter.Status)
	w.eq(`i.priority`, filter.Priority != nil, filter.Priority)
	w.eq(`i.sprint_id`, filter.SprintID != nil, filter.SprintID)
	w.raw(filter.InBacklog != nil && *filter.InBacklog, `i.sprint_id IS NULL`)
	w.eq(`i.triaged`, filter.Triaged != nil, filter.Triaged)
	w.eq(`i.refined`, filter.Refined != nil, filter.Refined)
	w.clause(`EXISTS (SELECT 1 FROM issue_assignees ia WHERE ia.issue_id = i.id AND ia.user_id=$%d)`, filter.AssigneeID != nil, filter.AssigneeID)
	w.clause(`to_tsvector('english', i.title || ' ' || coalesce(i.description, '')) @@ plainto_tsquery('english', $%d)`, filter.Text != nil && *filter.Text != "", filter.Text)
	w.eq(`i.type`, filter.Type != nil, filter.Type)
	w.eq(`i.parent_id`, filter.ParentID != nil, filter.ParentID)
	w.raw(filter.HasNoParent != nil && *filter.HasNoParent, `i.parent_id IS NULL`)
	w.clause(`EXISTS (SELECT 1 FROM issue_labels il WHERE il.issue_id = i.id AND il.label_id=$%d)`, filter.LabelID != nil, filter.LabelID)
	w.clause(`NOT EXISTS (SELECT 1 FROM issue_labels il WHERE il.issue_id = i.id AND il.label_id=$%d)`, filter.ExcludeLabelID != nil, filter.ExcludeLabelID)
	w.eq(`i.on_hold`, filter.OnHold != nil, filter.OnHold)
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
		i.priority, i.estimate, i.reporter_id, i.owner_id, i.parent_id, i.milestone_id, i.sprint_id, i.sprint_carry_count,
		i.triaged, i.refined, i.visibility, i.customer_email, i.customer_name, i.customer_token,
		i."rank", i.cancel_reason, i.checklist, i.created_at, i.updated_at, i.version,
		COALESCE((SELECT array_agg(user_id) FROM issue_assignees WHERE issue_id=i.id), '{}'),
		COALESCE((SELECT array_agg(label_id) FROM issue_labels WHERE issue_id=i.id), '{}')
	FROM issues i`

func scanIssue(row *sql.Row) (*models.Issue, error) {
	var id, projectID uuid.UUID
	var number int
	var issueType models.IssueType
	var title string
	var desc, holdReason, holdNote, customerEmail, customerName, rankStr, cancelReason sql.NullString
	var status models.IssueStatus
	var onHold bool
	var holdSince sql.NullTime
	var priority models.IssuePriority
	var estimate models.IssueEstimate
	var reporterID, ownerID, parentID, milestoneID, sprintID uuid.NullUUID
	var customerToken uuid.NullUUID
	var sprintCarryCount int
	var triaged, refined bool
	var visibility models.IssueVisibility
	var checklistJSON []byte
	var createdAt, updatedAt time.Time
	var version int
	var assigneeArr, labelArr []byte

	err := row.Scan(
		&id, &projectID, &number, &issueType, &title, &desc, &status,
		&onHold, &holdReason, &holdSince, &holdNote,
		&priority, &estimate, &reporterID, &ownerID, &parentID, &milestoneID, &sprintID, &sprintCarryCount,
		&triaged, &refined, &visibility, &customerEmail, &customerName, &customerToken,
		&rankStr, &cancelReason, &checklistJSON, &createdAt, &updatedAt, &version,
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
		reporterID, ownerID, parentID, milestoneID, sprintID,
		sprintCarryCount, triaged, refined, visibility,
		rankStr, cancelReason, checklistJSON, createdAt, updatedAt, version,
		assigneeArr, labelArr,
	), nil
}

func scanIssueRow(rows *sql.Rows) (*models.Issue, error) {
	var id, projectID uuid.UUID
	var number int
	var issueType models.IssueType
	var title string
	var desc, holdReason, holdNote, customerEmail, customerName, rankStr, cancelReason sql.NullString
	var status models.IssueStatus
	var onHold bool
	var holdSince sql.NullTime
	var priority models.IssuePriority
	var estimate models.IssueEstimate
	var reporterID, ownerID, parentID, milestoneID, sprintID uuid.NullUUID
	var customerToken uuid.NullUUID
	var sprintCarryCount int
	var triaged, refined bool
	var visibility models.IssueVisibility
	var checklistJSON []byte
	var createdAt, updatedAt time.Time
	var version int
	var assigneeArr, labelArr []byte

	err := rows.Scan(
		&id, &projectID, &number, &issueType, &title, &desc, &status,
		&onHold, &holdReason, &holdSince, &holdNote,
		&priority, &estimate, &reporterID, &ownerID, &parentID, &milestoneID, &sprintID, &sprintCarryCount,
		&triaged, &refined, &visibility, &customerEmail, &customerName, &customerToken,
		&rankStr, &cancelReason, &checklistJSON, &createdAt, &updatedAt, &version,
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
		reporterID, ownerID, parentID, milestoneID, sprintID,
		sprintCarryCount, triaged, refined, visibility,
		rankStr, cancelReason, checklistJSON, createdAt, updatedAt, version,
		assigneeArr, labelArr,
	), nil
}

func buildIssue(
	id, projectID uuid.UUID, number int, issueType models.IssueType, title string,
	desc, holdReason, holdNote, customerEmail, customerName sql.NullString,
	status models.IssueStatus, onHold bool, holdSince sql.NullTime,
	customerToken uuid.NullUUID,
	priority models.IssuePriority, estimate models.IssueEstimate,
	reporterID, ownerID, parentID, milestoneID, sprintID uuid.NullUUID,
	sprintCarryCount int, triaged bool, refined bool, visibility models.IssueVisibility,
	rankStr, cancelReason sql.NullString,
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
	var ownerIDPtr *uuid.UUID
	if ownerID.Valid {
		ownerIDPtr = &ownerID.UUID
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
	var cancelReasonPtr *string
	if cancelReason.Valid {
		cancelReasonPtr = &cancelReason.String
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
		reporterIDPtr, ownerIDPtr, parentIDPtr, milestoneIDPtr, sprintIDPtr,
		sprintCarryCount, triaged, refined, visibility,
		customerEmailPtr, customerNamePtr, customerTokenPtr,
		rankPtr, cancelReasonPtr, checklist, assignees, labels, nil,
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
