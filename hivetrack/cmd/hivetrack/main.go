package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/The127/ioc"
	"github.com/the127/hivetrack/internal/config"
	"github.com/the127/hivetrack/internal/database"
	"github.com/the127/hivetrack/internal/server"
	"github.com/the127/hivetrack/internal/setup"
	"go.uber.org/zap"
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
	defer db.Close()

	// Run migrations
	logger.Info("running migrations")
	if err := database.Migrate(db); err != nil {
		logger.Fatal("running migrations", zap.Error(err))
	}
	logger.Info("migrations complete")

	// Wire DI
	dc := ioc.NewDependencyCollection()
	setup.Database(dc, db)
	setup.Services(dc, cfg)
	setup.Mediator(dc)

	dp := dc.BuildProvider()

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
