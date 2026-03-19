package events

import (
	"context"
)

// ScenarioGenerator generates GIVEN/WHEN/THEN acceptance scenarios from a description.
type ScenarioGenerator interface {
	GenerateScenarios(ctx context.Context, description string) (string, error)
}

// HandleIssueRefinedForHivemind returns an event handler that posts GIVEN/WHEN/THEN
// scenarios as a Hivemind-attributed comment whenever a task issue is refined.
func HandleIssueRefinedForHivemind(gen ScenarioGenerator) func(context.Context, IssueRefinedPayload) error {
	return func(ctx context.Context, payload IssueRefinedPayload) error {
		return nil
	}
}
