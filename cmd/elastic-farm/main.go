package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/hiepnv90/elastic-farm/internal/config"
	"github.com/hiepnv90/elastic-farm/pkg/common"
	"github.com/hiepnv90/elastic-farm/pkg/graphql"
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

	fmt.Println(*cfg)

	client = graphql.New(cfg.GraphQL, nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	monitorPosition(ctx, cfg.Position)
}

func monitorPosition(ctx context.Context, pos string) {
	ticker := time.NewTicker(time.Second)

	posState := positionState{
		Amount0: common.Big0,
		Amount1: common.Big0,
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stop monitoring position:", pos)
			return
		case <-ticker.C:
			positions, err := client.GetPositions([]string{pos})
			if err != nil {
				fmt.Println("Fail to get position data:", err)
				continue
			}
			if len(positions) == 0 {
				continue
			}

			posData := positions[0]
			currentTick, err := strconv.Atoi(posData.Pool.Tick)
			if err != nil {
				fmt.Println("Fail to parse current tick:", err)
				continue
			}

			tickLower, err := strconv.Atoi(posData.TickLower.TickIdx)
			if err != nil {
				fmt.Println("Fail to parse tick lower:", err)
				continue
			}

			tickUpper, err := strconv.Atoi(posData.TickUpper.TickIdx)
			if err != nil {
				fmt.Println("Fail to parse tick upper:", err)
				continue
			}

			sqrtPrice := common.NewBigIntFromString(posData.Pool.SqrtPrice, 10)
			liquidity := common.NewBigIntFromString(posData.Liquidity, 10)

			token0Decimals, err := strconv.Atoi(posData.Pool.Token0.Decimals)
			if err != nil {
				fmt.Println("Fail to parse token decimals:", err)
				continue
			}

			token1Decimals, err := strconv.Atoi(posData.Pool.Token1.Decimals)
			if err != nil {
				fmt.Println("Fail to parse token decimals:", err)
				continue
			}

			amount0, amount1 := common.ExtractLiquidity(currentTick, tickLower, tickUpper, sqrtPrice, liquidity)
			newPosState := positionState{
				Amount0: amount0,
				Amount1: amount1,
			}
			if !posState.Equal(newPosState) {
				posState = newPosState
				fmt.Printf(
					"Position: %s %s %s %s\n",
					common.FormatAmount(amount0, token0Decimals),
					posData.Pool.Token0.Symbol,
					common.FormatAmount(amount1, token1Decimals),
					posData.Pool.Token1.Symbol,
				)
			}
		}
	}
}

type positionState struct {
	Amount0 *big.Int
	Amount1 *big.Int
}

func (p positionState) Equal(p1 positionState) bool {
	return p.Amount0.Cmp(p1.Amount0) == 0 && p.Amount1.Cmp(p1.Amount1) == 0
}
