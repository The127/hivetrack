package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

// ErrVagueCriteria is the sentinel error for vague or non-actionable acceptance criteria.
// Use NewVagueCriteriaError to construct an error that carries a specific gap explanation.
var ErrVagueCriteria = errors.New("acceptance criteria are too vague to generate scenarios")

// VagueCriteriaError wraps ErrVagueCriteria and carries a structured explanation of
// which acceptance criteria are insufficient and what information is missing.
type VagueCriteriaError struct {
	Explanation string
}

func (e *VagueCriteriaError) Error() string {
	return ErrVagueCriteria.Error() + ": " + e.Explanation
}

func (e *VagueCriteriaError) Is(target error) bool {
	return target == ErrVagueCriteria
}

// NewVagueCriteriaError constructs a VagueCriteriaError with the given explanation of
// which criteria are insufficient and what information is missing.
func NewVagueCriteriaError(explanation string) *VagueCriteriaError {
	return &VagueCriteriaError{Explanation: explanation}
}

// ScenarioGenerator generates GIVEN/WHEN/THEN acceptance scenarios from a description.
type ScenarioGenerator interface {
	GenerateScenarios(ctx context.Context, description string) (string, error)
}

const (
	hivemindEmail = "hivemind@hivetrack.internal"
	hivemindName  = "Hivemind"
)

// HandleIssueRefinedForHivemind returns an event handler that posts GIVEN/WHEN/THEN
// scenarios as a Hivemind-attributed comment whenever a task issue is refined.
func HandleIssueRefinedForHivemind(gen ScenarioGenerator) func(context.Context, IssueRefinedPayload) error {
	return func(ctx context.Context, payload IssueRefinedPayload) error {
		db := repositories.GetDbContext(ctx)

		issue, err := db.Issues().GetByID(ctx, payload.IssueID)
		if err != nil {
			return err
		}
		if issue == nil {
			return nil
		}
		if issue.GetType() != models.IssueTypeTask {
			return nil
		}

		desc := issue.GetDescription()

		var body string
		if desc == nil || *desc == "" {
			if err := revertRefined(ctx, db, issue); err != nil {
				return err
			}
			body = "Hivemind could not generate acceptance scenarios because no description or acceptance criteria were found. Please add a description with clear, testable criteria and re-refine the issue."
		} else {
			scenarios, genErr := gen.GenerateScenarios(ctx, *desc)
			if errors.Is(genErr, ErrVagueCriteria) {
				if err := revertRefined(ctx, db, issue); err != nil {
					return err
				}
				body = vagueCriteriaGapBody(genErr)
			} else if genErr != nil {
				return genErr
			} else {
				body = scenarios
			}
		}

		email := hivemindEmail
		name := hivemindName
		comment := models.NewComment(issue.GetId(), nil, &email, &name, body)
		db.Comments().Insert(comment)

		return db.SaveChanges(ctx)
	}
}

// vagueCriteriaGapBody returns a gap comment body for an ErrVagueCriteria error.
// If the error is a VagueCriteriaError with a non-empty Explanation, the explanation
// is embedded in the body so the author knows which criteria are insufficient.
func vagueCriteriaGapBody(err error) string {
	var vce *VagueCriteriaError
	if errors.As(err, &vce) && vce.Explanation != "" {
		return "Hivemind could not generate acceptance scenarios because the criteria are too vague or non-actionable.\n\n" +
			vce.Explanation +
			"\n\nPlease address the gaps above and re-refine the issue."
	}
	return "Hivemind could not generate acceptance scenarios because the criteria are too vague or non-actionable. Please add specific, testable criteria and re-refine the issue."
}

// revertRefined marks the issue as not refined and enqueues an issue.unrefined outbox event.
func revertRefined(ctx context.Context, db repositories.DbContext, issue *models.Issue) error {
	issue.SetRefined(false)
	db.Issues().Update(issue)
	payload, err := json.Marshal(IssueUnrefinedPayload{IssueID: issue.GetId(), ProjectID: issue.GetProjectID()})
	if err != nil {
		return fmt.Errorf("marshaling issue.unrefined payload: %w", err)
	}
	if err := db.Outbox().Enqueue(ctx, EventTypeIssueUnrefined, payload); err != nil {
		return fmt.Errorf("enqueueing issue.unrefined event: %w", err)
	}
	return nil
}
