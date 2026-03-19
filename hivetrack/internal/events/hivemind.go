package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

// ErrVagueCriteria is returned by ScenarioGenerator when the acceptance criteria
// are too vague or non-actionable to generate meaningful scenarios.
var ErrVagueCriteria = errors.New("acceptance criteria are too vague to generate scenarios")

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
			scenarios, err := gen.GenerateScenarios(ctx, *desc)
			if errors.Is(err, ErrVagueCriteria) {
				if err := revertRefined(ctx, db, issue); err != nil {
					return err
				}
				body = "Hivemind could not generate acceptance scenarios because the criteria are too vague or non-actionable. Please add specific, testable criteria and re-refine the issue."
			} else if err != nil {
				return err
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
