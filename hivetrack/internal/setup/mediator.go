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
	mediatr.RegisterHandler(m, commands.HandleAddProjectMember)
	mediatr.RegisterHandler(m, commands.HandleRemoveProjectMember)
	mediatr.RegisterHandler(m, commands.HandleCreateIssue)
	mediatr.RegisterHandler(m, commands.HandleUpdateIssue)
	mediatr.RegisterHandler(m, commands.HandleDeleteIssue)
	mediatr.RegisterHandler(m, commands.HandleTriageIssue)
	mediatr.RegisterHandler(m, commands.HandleCreateSprint)
	mediatr.RegisterHandler(m, commands.HandleUpdateSprint)
	mediatr.RegisterHandler(m, commands.HandleDeleteSprint)
	mediatr.RegisterHandler(m, commands.HandleAddChecklistItem)
	mediatr.RegisterHandler(m, commands.HandleUpdateChecklistItem)
	mediatr.RegisterHandler(m, commands.HandleRemoveChecklistItem)
	mediatr.RegisterHandler(m, commands.HandleCreateLabel)
	mediatr.RegisterHandler(m, commands.HandleUpdateLabel)
	mediatr.RegisterHandler(m, commands.HandleDeleteLabel)
	mediatr.RegisterHandler(m, commands.HandleCreateComment)
	mediatr.RegisterHandler(m, commands.HandleUpdateComment)
	mediatr.RegisterHandler(m, commands.HandleDeleteComment)

	// Queries
	mediatr.RegisterHandler(m, queries.HandleGetUsers)
	mediatr.RegisterHandler(m, queries.HandleGetProjects)
	mediatr.RegisterHandler(m, queries.HandleGetProject)
	mediatr.RegisterHandler(m, queries.HandleGetIssues)
	mediatr.RegisterHandler(m, queries.HandleGetIssue)
	mediatr.RegisterHandler(m, queries.HandleGetMyIssues)
	mediatr.RegisterHandler(m, queries.HandleGetSprints)
	mediatr.RegisterHandler(m, queries.HandleGetMilestones)
	mediatr.RegisterHandler(m, queries.HandleGetLabels)
	mediatr.RegisterHandler(m, queries.HandleGetComments)

	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) mediatr.Mediator {
		return m
	})
}
