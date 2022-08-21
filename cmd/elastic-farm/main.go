package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/hiepnv90/elastic-farm/internal/app"
	"github.com/hiepnv90/elastic-farm/internal/config"
	"github.com/hiepnv90/elastic-farm/pkg/common"
	"github.com/hiepnv90/elastic-farm/pkg/graphql"
	"github.com/hiepnv90/elastic-farm/pkg/position"
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

	zap.S().Infow("Start monitoring farming position", "cfg", cfg)

	zap.S().Infow("Create new client for GraphQL", "baseURL", cfg.GraphQL)
	client = graphql.New(cfg.GraphQL, nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	zap.S().Infow("Start monitoring farming position", "position", cfg.Position)
	monitorPosition(ctx, cfg.Position)
}

func setupLogger(debug bool) *zap.Logger {
	logLevel := zap.InfoLevel
	if debug {
		logLevel = zap.DebugLevel
	}

	return app.NewLogger(logLevel)
}

func monitorPosition(ctx context.Context, pos string) {
	l := zap.S().With("pos", pos)

	ticker := time.NewTicker(time.Second)

	posInfo, err := getPosition(client, pos)
	if err != nil {
		l.Fatalw("Fail to get position information", "error", err)
	}
	l.Infow("Position information", "info", posInfo.String())

	for {
		select {
		case <-ctx.Done():
			l.Infow("Stop monitoring position")
			return
		case <-ticker.C:
			newPosInfo, err := getPosition(client, pos)
			if err != nil {
				l.Errorw("Fail to get position information", "error", err)
				continue
			}
			if !posInfo.Equal(newPosInfo) {
				posInfo = newPosInfo
				l.Infow("Position information", "info", posInfo.String())
			}
		}
	}
}

func getPosition(client *graphql.Client, pos string) (position.Position, error) {
	l := zap.S().With("pos", pos)

	l.Debugw("Get position liquidity")

	positions, err := client.GetPositions([]string{pos})
	if err != nil {
		l.Errorw("Fail to get position liquidity", "error", err)
		return position.Position{}, err
	}
	if len(positions) == 0 {
		l.Warnw("Position not found")
		return position.Position{}, errors.New("position not found")
	}

	posData := positions[0]
	l.Debugw("Position information", "posInfo", posData)

	currentTick, err := strconv.Atoi(posData.Pool.Tick)
	if err != nil {
		l.Errorw("Fail to parse current tick", "tick", posData.Pool.Tick, "error", err)
		return position.Position{}, err
	}

	tickLower, err := strconv.Atoi(posData.TickLower.TickIdx)
	if err != nil {
		l.Errorw("Fail to parse tick lower", "tick", posData.TickLower.TickIdx, "error", err)
		return position.Position{}, err
	}

	tickUpper, err := strconv.Atoi(posData.TickUpper.TickIdx)
	if err != nil {
		l.Errorw("Fail to parse tick upper", "tick", posData.TickUpper.TickIdx, "error", err)
		return position.Position{}, err
	}

	sqrtPrice := common.NewBigIntFromString(posData.Pool.SqrtPrice, 10)
	liquidity := common.NewBigIntFromString(posData.Liquidity, 10)

	token0Decimals, err := strconv.Atoi(posData.Pool.Token0.Decimals)
	if err != nil {
		l.Errorw("Fail to parse token0 decimals", "decimals", posData.Pool.Token0.Decimals, "error", err)
		return position.Position{}, err
	}

	token1Decimals, err := strconv.Atoi(posData.Pool.Token1.Decimals)
	if err != nil {
		l.Errorw("Fail to parse token1 decimals", "decimals", posData.Pool.Token1.Decimals, "error", err)
		return position.Position{}, err
	}

	amount0, amount1 := common.ExtractLiquidity(currentTick, tickLower, tickUpper, sqrtPrice, liquidity)
	return position.Position{
		Token0: common.Token{
			Amount:   amount0,
			Symbol:   posData.Pool.Token0.Symbol,
			Decimals: token0Decimals,
		},
		Token1: common.Token{
			Amount:   amount1,
			Symbol:   posData.Pool.Token1.Symbol,
			Decimals: token1Decimals,
		},
	}, nil
}
