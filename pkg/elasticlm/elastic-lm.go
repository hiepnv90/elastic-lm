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
	"github.com/hiepnv90/elastic-lm/pkg/models"
	"github.com/hiepnv90/elastic-lm/pkg/position"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var bps = big.NewInt(10000)

type ElasticLM struct {
	interval           time.Duration
	positionIDs        []string
	amountThresholdBps *big.Int
	positionMap        map[string]position.Position
	symbolInfoMap      map[string]futures.Symbol

	db      *gorm.DB
	client  *graphql.Client
	bclient *binance.Client
	logger  *zap.SugaredLogger
}

func New(
	db *gorm.DB,
	client *graphql.Client,
	bclient *binance.Client,
	positionIDs []string,
	amountThresholdBps int,
	interval time.Duration,
) *ElasticLM {
	return &ElasticLM{
		interval:           interval,
		positionIDs:        positionIDs,
		amountThresholdBps: big.NewInt(int64(amountThresholdBps)),
		positionMap:        make(map[string]position.Position),
		symbolInfoMap:      make(map[string]futures.Symbol),
		db:                 db,
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

	err := e.loadPositions()
	if err != nil {
		l.Errorw("Fail to load saved positions from database", "error", err)
		return err
	}

	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	err = e.updatePositions(ctx, isHedge)
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

	err = e.savePositions()
	if err != nil {
		l.Warnw("Fail to save positions into database", "error", err)
	}

	return nil
}

func (e *ElasticLM) updatePosition(newPosInfo position.Position, isHedge bool) error {
	l := e.logger

	posInfo, ok := e.positionMap[newPosInfo.ID]
	if ok {
		newPosInfo.HedgedAmount0 = posInfo.HedgedAmount0
		newPosInfo.HedgedAmount1 = posInfo.HedgedAmount1
		if posInfo.Equal(newPosInfo) {
			return nil
		}
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
			newPosInfo.HedgedAmount0 = amount0
			newPosInfo.HedgedAmount1 = amount1
			e.positionMap[newPosInfo.ID] = newPosInfo
		}
		return nil
	}

	// Check deltaAmount threshold for hedging base on token0.
	absThreshold := common.BigDiv(common.BigMul(posInfo.MaxAmount0, e.amountThresholdBps), bps)
	token0 := newPosInfo.Token0
	token0.Amount = common.BigSub(token0.Amount, posInfo.HedgedAmount0)
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

	// Hedge for token0 delta
	amount0, err := e.hedgeToken(token0)
	if err != nil {
		l.Warnw("Fail to hedge for token", "token", token0, "error", err)
	}
	newPosInfo.HedgedAmount0 = common.BigAdd(posInfo.HedgedAmount0, amount0)

	// Hedge for token1 delta
	token1 := newPosInfo.Token1
	token1.Amount = common.BigSub(token1.Amount, posInfo.HedgedAmount1)
	amount1, err := e.hedgeToken(token1)
	if err != nil {
		l.Warnw("Fail to hedge for token", "token", token1, "error", err)
	}
	newPosInfo.HedgedAmount1 = common.BigAdd(posInfo.HedgedAmount1, amount1)

	l.Infow("Update position hedged amounts", "newPosInfo", newPosInfo)
	e.positionMap[posInfo.ID] = newPosInfo

	return nil
}

func (e *ElasticLM) hedgeToken(token common.Token) (*big.Int, error) {
	if token.IsStable() {
		return token.Amount, nil
	}

	symbol := token.GetBinancePerpetualSymbol()
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
			TickLower:  tickLower,
			TickUpper:  tickUpper,
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

func (e *ElasticLM) loadPositions() error {
	l := e.logger

	l.Infow("Load open positions from database")
	var positions []models.Position
	err := e.db.Where("liquidity > 0").Find(&positions).Error
	if err != nil {
		l.Errorw("Fail to get positions from database", "error", err)
		return err
	}

	for _, pos := range positions {
		liquidity := common.NewBigIntFromString(pos.Liquidity, 10)
		amount0 := common.NewBigIntFromString(pos.Amount0, 10)
		amount1 := common.NewBigIntFromString(pos.Amount1, 10)
		hedgedAmount0 := common.NewBigIntFromString(pos.HedgedAmount0, 10)
		hedgedAmount1 := common.NewBigIntFromString(pos.HedgedAmount1, 10)

		lowerSqrtPrice := common.GetSqrtRatioAtTick(pos.TickLower)
		upperSqrtPrice := common.GetSqrtRatioAtTick(pos.TickUpper)
		maxAmount0 := common.CalculateAmount0(lowerSqrtPrice, upperSqrtPrice, liquidity)
		maxAmount1 := common.CalculateAmount1(lowerSqrtPrice, upperSqrtPrice, liquidity)
		e.positionMap[pos.ID] = position.Position{
			ID:            pos.ID,
			Liquidity:     liquidity,
			TickLower:     pos.TickLower,
			TickUpper:     pos.TickUpper,
			MaxAmount0:    maxAmount0,
			MaxAmount1:    maxAmount1,
			HedgedAmount0: hedgedAmount0,
			HedgedAmount1: hedgedAmount1,
			Token0: common.Token{
				Symbol:   pos.Symbol0,
				Decimals: pos.Decimals0,
				Amount:   amount0,
			},
			Token1: common.Token{
				Symbol:   pos.Symbol1,
				Decimals: pos.Decimals1,
				Amount:   amount1,
			},
		}
	}

	return nil
}

func (e *ElasticLM) savePositions() error {
	positions := make([]models.Position, 0, len(e.positionMap))
	for _, pos := range e.positionMap {
		positions = append(positions, models.Position{
			ID:            pos.ID,
			Liquidity:     pos.Liquidity.String(),
			TickLower:     pos.TickLower,
			TickUpper:     pos.TickUpper,
			Symbol0:       pos.Token0.Symbol,
			Amount0:       pos.Token0.Amount.String(),
			Decimals0:     pos.Token0.Decimals,
			HedgedAmount0: pos.HedgedAmount0.String(),
			Symbol1:       pos.Token1.Symbol,
			Amount1:       pos.Token1.Amount.String(),
			Decimals1:     pos.Token1.Decimals,
			HedgedAmount1: pos.HedgedAmount1.String(),
			UpdatedAt:     time.Now(),
		})
	}

	return e.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&positions).Error
}
