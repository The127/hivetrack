package setup

import (
	"github.com/The127/ioc"

	"github.com/the127/hivetrack/internal/events"
)

// Events registers the RefinementBroker as a singleton and returns it so
// callers (Mediator, Nats) can wire its Publish method into command handlers
// and the NATS subscriber.
func Events(dc *ioc.DependencyCollection) *events.RefinementBroker {
	broker := events.NewRefinementBroker()
	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) *events.RefinementBroker {
		return broker
	})
	return broker
}
