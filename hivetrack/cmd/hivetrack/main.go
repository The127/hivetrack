package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/The127/ioc"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"github.com/the127/hivetrack/internal/config"
	"github.com/the127/hivetrack/internal/database"
	"github.com/the127/hivetrack/internal/infrastructure"
	"github.com/the127/hivetrack/internal/server"
	"github.com/the127/hivetrack/internal/setup"
)

func main() {
	// Load config
	configPath := "config.yaml"
	if path := os.Getenv("HIVETRACK_CONFIG"); path != "" {
		configPath = path
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	// Init logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("creating logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	// Connect to database
	logger.Info("connecting to database")
	db, err := database.Open(cfg.Database.URL)
	if err != nil {
		logger.Fatal("connecting to database", zap.Error(err))
	}
	defer func() { _ = db.Close() }()

	// Run migrations
	logger.Info("running migrations")
	if err := database.Migrate(db); err != nil {
		logger.Fatal("running migrations", zap.Error(err))
	}
	logger.Info("migrations complete")

	// Connect to NATS (only when Hivemind is enabled)
	var nc *nats.Conn
	var js jetstream.JetStream
	if cfg.Hivemind.Enabled {
		logger.Info("connecting to NATS", zap.String("url", cfg.Hivemind.NatsURL))
		nc, err = nats.Connect(cfg.Hivemind.NatsURL)
		if err != nil {
			logger.Fatal("connecting to NATS", zap.Error(err))
		}
		defer nc.Close()

		js, err = jetstream.New(nc)
		if err != nil {
			logger.Fatal("creating JetStream context", zap.Error(err))
		}
		ctx := context.Background()
		for _, stream := range []string{"hivetrack-refinement", "hivemind-refinement"} {
			if _, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
				Name:     stream,
				Subjects: []string{stream + ".>"},
			}); err != nil {
				logger.Fatal("creating JetStream stream", zap.String("stream", stream), zap.Error(err))
			}
			logger.Info("JetStream stream ready", zap.String("stream", stream))
		}
	}

	// Wire DI
	dc := ioc.NewDependencyCollection()
	setup.Database(dc, db)
	setup.Services(dc, cfg)
	if nc != nil {
		pub := setup.Nats(dc, nc, js)
		setup.Mediator(dc, pub)
	} else {
		setup.Mediator(dc)
	}

	dp := dc.BuildProvider()

	// Start NATS subscriber (only when Hivemind is enabled)
	if nc != nil {
		subscriber := ioc.GetDependency[*infrastructure.NatsSubscriber](dp)
		if err := subscriber.Start(context.Background()); err != nil {
			logger.Fatal("starting NATS subscriber", zap.Error(err))
		}
		logger.Info("NATS refinement subscriber started")
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("starting server", zap.String("addr", addr))

	go func() {
		if err := server.Serve(dp); err != nil {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	// Wait for signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Info("shutting down")
}
