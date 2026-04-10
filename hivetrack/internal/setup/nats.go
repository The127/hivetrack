package setup

import (
	"database/sql"

	"github.com/The127/ioc"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/infrastructure"
	"github.com/the127/hivetrack/internal/repositories"
	"github.com/the127/hivetrack/internal/repositories/postgres"
)

// Nats registers NATS-related singletons and returns a RefinementPublisher
// for use in Mediator registration. Only call when hivemind is enabled.
func Nats(dc *ioc.DependencyCollection, nc *nats.Conn, js jetstream.JetStream) (commands.RefinementPublisher, *infrastructure.TokenBuffer) {
	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) *nats.Conn {
		return nc
	})

	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) jetstream.JetStream {
		return js
	})

	pub := infrastructure.NewNatsPublisher(js)

	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) *infrastructure.NatsPublisher {
		return pub
	})

	buf := infrastructure.NewTokenBuffer()

	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) *infrastructure.TokenBuffer {
		return buf
	})

	ioc.RegisterSingleton(dc, func(dp *ioc.DependencyProvider) *infrastructure.NatsSubscriber {
		logger := ioc.GetDependency[*zap.Logger](dp)
		sqlDB := ioc.GetDependency[*sql.DB](dp)
		newRepo := func() repositories.RefinementRepository {
			return postgres.NewDbContext(sqlDB).Refinements()
		}
		return infrastructure.NewNatsSubscriber(js, newRepo, logger, buf)
	})

	return &refinementPublisherAdapter{pub: pub}, buf
}
