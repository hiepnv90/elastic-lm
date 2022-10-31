package binance

import (
	"context"

	binance "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"go.uber.org/zap"
)

type Client struct {
	apiKey    string
	secretKey string

	spotClient   *binance.Client
	futureClient *futures.Client
	logger       *zap.SugaredLogger
}

func New(apiKey string, secretKey string) *Client {
	spotClient := binance.NewClient(apiKey, secretKey)
	futureClient := binance.NewFuturesClient(apiKey, secretKey)

	return &Client{
		apiKey:       apiKey,
		secretKey:    secretKey,
		spotClient:   spotClient,
		futureClient: futureClient,
		logger:       zap.S(),
	}
}

func (c *Client) GetExchangeInfo(ctx context.Context) (*futures.ExchangeInfo, error) {
	c.logger.Debugw("Get exchange information")

	exchangeInfo, err := c.futureClient.NewExchangeInfoService().Do(ctx)
	if err != nil {
		c.logger.Errorw("Fail to get exchange information", "error", err)
		return nil, err
	}

	return exchangeInfo, nil
}

func (c *Client) CreateFutureOrder(
	ctx context.Context,
	symbol string,
	quantity string,
	price string,
	side futures.SideType,
	orderType futures.OrderType,
	timeInForce futures.TimeInForceType,
	reduceOnly bool,
) (*futures.CreateOrderResponse, error) {
	c.logger.Infow(
		"Create futures's order",
		"symbol", symbol,
		"quantity", quantity,
		"price", price,
		"side", side,
		"type", orderType,
		"timeInForce", timeInForce,
		"reduceOnly", reduceOnly,
	)

	createOrderService := c.futureClient.
		NewCreateOrderService().
		Symbol(symbol).
		Quantity(quantity).
		Side(side).
		Type(orderType).
		ReduceOnly(reduceOnly)
	if orderType != futures.OrderTypeMarket {
		createOrderService = createOrderService.Price(price).TimeInForce(timeInForce)
	}

	resp, err := createOrderService.Do(ctx)
	if err != nil {
		c.logger.Errorw(
			"Fail to create future order",
			"symbol", symbol,
			"quantity", quantity,
			"price", price,
			"side", side,
			"type", orderType,
			"timeInForce", timeInForce,
			"reduceOnly", reduceOnly,
			"error", err,
		)
		return nil, err
	}

	return resp, nil
}

func (c *Client) ListenUserData(
	ctx context.Context,
	eventC chan *futures.WsUserDataEvent,
) (doneC, stopC chan struct{}, err error) {
	listenKey, err := c.futureClient.NewStartUserStreamService().Do(ctx)
	if err != nil {
		c.logger.Errorw("Fail to create listen key", "error", err)
		return
	}

	doneC, stopC, err = futures.WsUserDataServe(
		listenKey,
		func(event *futures.WsUserDataEvent) {
			eventC <- event
		},
		func(err error) {
			c.logger.Errorw("Listen user data error", "error", err)
		},
	)
	if err != nil {
		c.logger.Errorw("Fail to listen user data", "error", err)
		return
	}

	return
}
