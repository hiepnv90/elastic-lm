package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/hiepnv90/elastic-lm/internal/app"
	"github.com/hiepnv90/elastic-lm/internal/config"
	"github.com/hiepnv90/elastic-lm/pkg/elasticlm"
	"github.com/hiepnv90/elastic-lm/pkg/graphql"
	"go.uber.org/zap"
)

var (
	configFile = flag.String("config", "config.yaml", "Path to configuration file")

	cfg    *config.Config
	client *graphql.Client
)

func main() {
	flag.Parse()

	var err error
	cfg, err = config.FromFile(*configFile)
	if err != nil {
		panic(err)
	}

	logger := setupLogger(cfg.Debug)
	defer func() {
		_ = logger.Sync()
	}()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	zap.S().Infow("Start application", "cfg", cfg)

	zap.S().Infow("Create new client for GraphQL", "baseURL", cfg.GraphQL)
	client = graphql.New(cfg.GraphQL, nil)

	zap.S().Infow("Create new ElasticLM instance", "positions", cfg.Positions)
	elasticLM := elasticlm.New(client, cfg.Positions, time.Second)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err = elasticLM.Start(ctx)
	if err != nil {
		zap.S().Fatalw("Fail to monitor positions", "error", err)
	}
	zap.S().Infow("Stop application!")
}

func setupLogger(debug bool) *zap.Logger {
	logLevel := zap.InfoLevel
	if debug {
		logLevel = zap.DebugLevel
	}

	return app.NewLogger(logLevel)
}
