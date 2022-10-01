package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/hiepnv90/elastic-lm/internal/app"
	"github.com/hiepnv90/elastic-lm/internal/config"
	"github.com/hiepnv90/elastic-lm/pkg/binance"
	"github.com/hiepnv90/elastic-lm/pkg/elasticlm"
	"github.com/hiepnv90/elastic-lm/pkg/graphql"
	"github.com/hiepnv90/elastic-lm/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

	zap.S().Infow("Create new client for GraphQL", "baseURL", cfg.GraphQL)
	client = graphql.New(cfg.GraphQL, nil)

	zap.S().Infow("Create new binance's client")
	bclient := setupBinanceClient(cfg.Binance)

	zap.S().Infow("Setup database connection", "cfg", cfg.SQLite)
	db := setupDB(cfg.SQLite)

	zap.S().Infow("Create new ElasticLM instance", "positions", cfg.Positions)
	tokenInstrumentMap := make(map[string]string)
	for _, tokenInstrument := range cfg.Binance.Symbols {
		token := strings.ToUpper(tokenInstrument.Token)
		instrument := strings.ToUpper(tokenInstrument.Instrument)
		tokenInstrumentMap[token] = instrument
	}
	elasticLM := elasticlm.New(
		db, client, bclient, cfg.Positions, cfg.AmountThresholdBps, time.Second, tokenInstrumentMap,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err = elasticLM.Run(ctx)
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

func setupBinanceClient(cfg config.Binance) *binance.Client {
	if cfg.APIKey == "" || cfg.SecretKey == "" {
		return nil
	}
	return binance.New(cfg.APIKey, cfg.SecretKey)
}

func setupDB(cfg config.SQLite) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(cfg.DBName), &gorm.Config{})
	if err != nil {
		zap.S().Fatalw("Fail to open database connection", "cfg", cfg, "error", err)
	}

	err = models.AutoMigrate(db)
	if err != nil {
		zap.S().Fatalw("Fail to auto migrate database", "error", err)
	}

	return db
}
