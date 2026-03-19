package events

import (
	"context"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

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

		desc := issue.GetDescription()
		if desc == nil || *desc == "" {
			return nil
		}

		scenarios, err := gen.GenerateScenarios(ctx, *desc)
		if err != nil {
			return err
		}

		email := hivemindEmail
		name := hivemindName
		comment := models.NewComment(issue.GetId(), nil, &email, &name, scenarios)
		db.Comments().Insert(comment)

		return db.SaveChanges(ctx)
	}
}
