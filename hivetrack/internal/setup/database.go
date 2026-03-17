package setup

import (
	"context"
	"database/sql"

	"github.com/The127/ioc"
	"github.com/the127/hivetrack/internal/repositories"
	"github.com/the127/hivetrack/internal/repositories/postgres"
)

// Database registers the *sql.DB singleton and the scoped DbContext.
func Database(dc *ioc.DependencyCollection, db *sql.DB) {
	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) *sql.DB {
		return db
	})

	ioc.RegisterScoped(dc, func(dp *ioc.DependencyProvider) repositories.DbContext {
		sqlDB := ioc.GetDependency[*sql.DB](dp)
		return postgres.NewDbContext(sqlDB)
	})

	ioc.RegisterCloseHandler(dc, func(dbCtx repositories.DbContext) error {
		return dbCtx.Rollback(context.Background())
	})
}
