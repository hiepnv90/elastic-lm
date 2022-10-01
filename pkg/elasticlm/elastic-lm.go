package elasticlm

import (
	"context"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/hiepnv90/elastic-lm/pkg/binance"
	"github.com/hiepnv90/elastic-lm/pkg/common"
	"github.com/hiepnv90/elastic-lm/pkg/graphql"
	"github.com/hiepnv90/elastic-lm/pkg/position"
	"go.uber.org/zap"
)

var bps = big.NewInt(10000)

type ElasticLM struct {
	interval           time.Duration
	positionIDs        []string
	amountThresholdBps *big.Int
	positionMap        map[string]position.Position
	positionsSnapshot  map[string]position.Position
	symbolInfoMap      map[string]futures.Symbol
	tokenInstrumentMap map[string]string

	client  *graphql.Client
	bclient *binance.Client
	logger  *zap.SugaredLogger
}

func New(
	client *graphql.Client,
	bclient *binance.Client,
	positionIDs []string,
	amountThresholdBps int,
	interval time.Duration,
	tokenInstrumentMap map[string]string,
) *ElasticLM {
	return &ElasticLM{
		interval:           interval,
		positionIDs:        positionIDs,
		amountThresholdBps: big.NewInt(int64(amountThresholdBps)),
		positionMap:        make(map[string]position.Position),
		positionsSnapshot:  make(map[string]position.Position),
		symbolInfoMap:      make(map[string]futures.Symbol),
		tokenInstrumentMap: tokenInstrumentMap,
		client:             client,
		bclient:            bclient,
		logger:             zap.S(),
	}
}

func (e *ElasticLM) Run(ctx context.Context) error {
	l := e.logger.With("positions", e.positionIDs, "interval", e.interval)

	isHedge := e.bclient != nil
	l.Infow("Start monitoring positions", "isHedge", isHedge)

	if isHedge {
		l.Infow("Get exchange information")
		exchangeInfo, err := e.bclient.GetExchangeInfo(ctx)
		if err != nil {
			l.Errorw("Fail to get exchange information", "error", err)
			return err
		}
		for _, symbolInfo := range exchangeInfo.Symbols {
			e.symbolInfoMap[symbolInfo.Symbol] = symbolInfo
		}
	}

	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	err := e.updatePositions(ctx, isHedge)
	if err != nil {
		l.Errorw("Fail to update positions' information", "error", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			l.Infow("Stop monitoring positions")
			return nil
		case <-ticker.C:
			err = e.updatePositions(ctx, isHedge)
			if err != nil {
				l.Errorw("Fail to update positions' information", "error", err)
			}
		}
	}
}

func (e *ElasticLM) updatePositions(ctx context.Context, isHedge bool) error {
	l := e.logger

	posInfos, err := e.getPositions(ctx)
	if err != nil {
		l.Errorw("Fail to get positions' information", "positions", e.positionIDs, "error", err)
		return err
	}

	for _, posInfo := range posInfos {
		err = e.updatePosition(posInfo, isHedge)
		if err != nil {
			l.Warnw("Fail to update position information", "info", posInfo.String(), "error", err)
		}
	}

	return nil
}

func (e *ElasticLM) updatePosition(newPosInfo position.Position, isHedge bool) error {
	l := e.logger

	posInfo, ok := e.positionMap[newPosInfo.ID]
	if ok && posInfo.Equal(newPosInfo) {
		return nil
	}

	l.Infow("Update position's information", "info", newPosInfo.String())
	e.positionMap[newPosInfo.ID] = newPosInfo

	if !isHedge {
		return nil
	}

	if !ok {
		if e.amountThresholdBps.Cmp(bps) < 0 {
			amount0, err := e.hedgeToken(newPosInfo.Token0)
			if err != nil {
				l.Warnw("Fail to hedge for token", "token", newPosInfo.Token0.String(), "error", err)
			}

			amount1, err := e.hedgeToken(newPosInfo.Token1)
			if err != nil {
				l.Warnw("Fail to hedge for token", "token", newPosInfo.Token1.String(), "error", err)
			}

			newPosInfo.Token0.Amount = amount0
			newPosInfo.Token1.Amount = amount1
			e.positionsSnapshot[newPosInfo.ID] = newPosInfo
		} else {
			newPosInfo.Token0.Amount = common.Big0
			newPosInfo.Token1.Amount = common.Big0
			e.positionsSnapshot[newPosInfo.ID] = newPosInfo
		}
		return nil
	}

	// Check deltaAmount threshold for hedging base on token0.
	posSnapshot := e.positionsSnapshot[newPosInfo.ID]
	absThreshold := common.BigDiv(common.BigMul(posSnapshot.MaxAmount0, e.amountThresholdBps), bps)
	token0 := newPosInfo.Token0
	token0.Amount = common.BigSub(token0.Amount, posSnapshot.Token0.Amount)
	if common.BigAbs(token0.Amount).Cmp(absThreshold) <= 0 &&
		newPosInfo.Token0.Amount.Cmp(newPosInfo.MaxAmount0) < 0 &&
		newPosInfo.Token0.Amount.Cmp(common.Big0) > 0 {
		l.Infow(
			"Ignore hedging for small change of amount",
			"token", token0,
			"absThreshold", common.FormatAmount(absThreshold, token0.Decimals, 5),
		)
		return nil
	}

	if posSnapshot.Liquidity.Cmp(newPosInfo.Liquidity) != 0 {
		posSnapshot.Liquidity = newPosInfo.Liquidity
		posSnapshot.MaxAmount0 = newPosInfo.MaxAmount0
		posSnapshot.MaxAmount1 = newPosInfo.MaxAmount1
	}

	// Hedge for token0 delta
	amount0, err := e.hedgeToken(token0)
	if err != nil {
		l.Warnw("Fail to hedge for token", "token", token0, "error", err)
	}
	posSnapshot.Token0.Amount = common.BigAdd(posSnapshot.Token0.Amount, amount0)

	// Hedge for token1 delta
	token1 := newPosInfo.Token1
	token1.Amount = common.BigSub(token1.Amount, posSnapshot.Token1.Amount)
	amount1, err := e.hedgeToken(token1)
	if err != nil {
		l.Warnw("Fail to hedge for token", "token", token1, "error", err)
	}
	posSnapshot.Token1.Amount = common.BigAdd(posSnapshot.Token1.Amount, amount1)

	l.Infow("Update position snapshot", "snapshot", posSnapshot)
	e.positionsSnapshot[newPosInfo.ID] = posSnapshot

	return nil
}

func (e *ElasticLM) hedgeToken(token common.Token) (*big.Int, error) {
	if token.IsStable() {
		return token.Amount, nil
	}

	symbol := e.getBinancePerpetualSymbol(token)
	symbolInfo := e.symbolInfoMap[symbol]
	precision := symbolInfo.QuantityPrecision

	amount := token.Amount
	side := futures.SideTypeSell
	reduceOnly := false
	if amount.Cmp(common.Big0) < 0 {
		amount = common.BigNeg(amount)
		side = futures.SideTypeBuy
		reduceOnly = true
	}

	amount = common.RoundAmount(amount, token.Decimals, precision, common.RoundTypeFloor)
	if common.BigIsZero(amount) {
		return common.Big0, nil
	}

	e.logger.Infow(
		"Hedging for token",
		"token", token,
		"precision", precision,
		"roundAmount", common.FormatAmount(amount, token.Decimals, 5),
	)

	resp, err := e.bclient.CreateFutureOrder(
		context.Background(),
		symbol,
		common.FormatAmount(amount, token.Decimals, precision),
		"0",
		side,
		futures.OrderTypeMarket,
		futures.TimeInForceTypeGTC,
		reduceOnly,
	)
	if err != nil {
		if strings.Contains(err.Error(), "code=-4164") {
			return common.Big0, nil
		}
		e.logger.Errorw("Fail to create future order", "error", err)
		return common.Big0, err
	}

	e.logger.Infow("Successfully create futures' order", "resp", resp)

	if side == futures.SideTypeBuy {
		amount = common.BigNeg(amount)
	}

	return amount, nil
}

func (e *ElasticLM) getPositions(ctx context.Context) ([]position.Position, error) {
	l := e.logger

	l.Debugw("Get positions' information", "positions", e.positionIDs)

	positions, err := e.client.GetPositions(e.positionIDs)
	if err != nil {
		l.Errorw("Fail to get position liquidity", "positions", e.positionIDs, "error", err)
		return nil, err
	}

	res := make([]position.Position, 0, len(positions))
	for _, posData := range positions {
		l.Debugw("Position information", "posInfo", posData)

		currentTick, err := strconv.Atoi(posData.Pool.Tick)
		if err != nil {
			l.Errorw("Fail to parse current tick", "tick", posData.Pool.Tick, "error", err)
			return nil, err
		}

		tickLower, err := strconv.Atoi(posData.TickLower.TickIdx)
		if err != nil {
			l.Errorw("Fail to parse tick lower", "tick", posData.TickLower.TickIdx, "error", err)
			return nil, err
		}

		tickUpper, err := strconv.Atoi(posData.TickUpper.TickIdx)
		if err != nil {
			l.Errorw("Fail to parse tick upper", "tick", posData.TickUpper.TickIdx, "error", err)
			return nil, err
		}

		sqrtPrice := common.NewBigIntFromString(posData.Pool.SqrtPrice, 10)
		liquidity := common.NewBigIntFromString(posData.Liquidity, 10)

		token0Decimals, err := strconv.Atoi(posData.Pool.Token0.Decimals)
		if err != nil {
			l.Errorw("Fail to parse token0 decimals", "decimals", posData.Pool.Token0.Decimals, "error", err)
			return nil, err
		}

		token1Decimals, err := strconv.Atoi(posData.Pool.Token1.Decimals)
		if err != nil {
			l.Errorw("Fail to parse token1 decimals", "decimals", posData.Pool.Token1.Decimals, "error", err)
			return nil, err
		}

		lowerSqrtPrice := common.GetSqrtRatioAtTick(tickLower)
		upperSqrtPrice := common.GetSqrtRatioAtTick(tickUpper)
		maxAmount0 := common.CalculateAmount0(lowerSqrtPrice, upperSqrtPrice, liquidity)
		maxAmount1 := common.CalculateAmount1(lowerSqrtPrice, upperSqrtPrice, liquidity)

		amount0, amount1 := common.ExtractLiquidity(currentTick, tickLower, tickUpper, sqrtPrice, liquidity)
		res = append(res, position.Position{
			ID:         posData.ID,
			Liquidity:  liquidity,
			MaxAmount0: maxAmount0,
			MaxAmount1: maxAmount1,
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
		})
	}

	return res, nil
}

func (e *ElasticLM) getBinancePerpetualSymbol(token common.Token) string {
	symbol, ok := e.tokenInstrumentMap[strings.ToUpper(token.Symbol)]
	if ok {
		return symbol
	}

	return token.GetBinancePerpetualSymbol()
}
