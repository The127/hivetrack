package postgres

// Entity type constants used by the change tracker to dispatch SaveChanges.
const (
	projectEntityType = iota + 1
	issueEntityType
	sprintEntityType
	milestoneEntityType
	labelEntityType
	commentEntityType
)
