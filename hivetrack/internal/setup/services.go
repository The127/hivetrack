package setup

import (
	"github.com/The127/ioc"
	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/config"
	"go.uber.org/zap"
)

// Services registers config, logger, and OIDC verifier as singletons.
// Must be called after Database() so that *sql.DB is not re-registered.
func Services(dc *ioc.DependencyCollection, cfg *config.Config) {
	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) *config.Config {
		return cfg
	})

	ioc.RegisterSingleton(dc, func(_ *ioc.DependencyProvider) *zap.Logger {
		logger, _ := zap.NewProduction()
		return logger
	})

	ioc.RegisterSingleton(dc, func(dp *ioc.DependencyProvider) *authentication.OIDCVerifier {
		cfg := ioc.GetDependency[*config.Config](dp)
		return authentication.NewOIDCVerifier(cfg.OIDC.Authority, cfg.OIDC.ClientID)
	})
}
