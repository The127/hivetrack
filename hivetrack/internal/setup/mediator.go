package setup

import (
	"github.com/The127/ioc"
	"github.com/The127/mediatr"
	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/queries"
)

// Mediator registers all command and query handlers and the mediator singleton.
func Mediator(dc *ioc.DependencyCollection) {
	m := mediatr.NewMediator()

	// Commands
	mediatr.RegisterHandler(m, commands.HandleCreateProject)
	mediatr.RegisterHandler(m, commands.HandleUpdateProject)
	mediatr.RegisterHandler(m, commands.HandleDeleteProject)
	mediatr.RegisterHandler(m, commands.HandleCreateIssue)
	mediatr.RegisterHandler(m, commands.HandleUpdateIssue)
	mediatr.RegisterHandler(m, commands.HandleDeleteIssue)
	mediatr.RegisterHandler(m, commands.HandleTriageIssue)
	mediatr.RegisterHandler(m, commands.HandleCreateSprint)
	mediatr.RegisterHandler(m, commands.HandleUpdateSprint)

	// Queries
	mediatr.RegisterHandler(m, queries.HandleGetProjects)
	mediatr.RegisterHandler(m, queries.HandleGetProject)
	mediatr.RegisterHandler(m, queries.HandleGetIssues)
	mediatr.RegisterHandler(m, queries.HandleGetIssue)
	mediatr.RegisterHandler(m, queries.HandleGetMyIssues)
	mediatr.RegisterHandler(m, queries.HandleGetSprints)
	mediatr.RegisterHandler(m, queries.HandleGetMilestones)
	mediatr.RegisterHandler(m, queries.HandleGetLabels)

	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) mediatr.Mediator {
		return m
	})
}
