package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	baseURL string

	httpClient *http.Client
}

func New(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *Client) Do(url string, method string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if method == http.MethodPost && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *Client) Post(url string, body interface{}) (*http.Response, error) {
	var bodyR io.Reader
	if body != nil {
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(body)
		if err != nil {
			return nil, err
		}
		bodyR = &buf
	}

	return c.Do(url, http.MethodPost, bodyR)
}

type Token struct {
	Symbol   string
	Decimals string
}

type Pool struct {
	SqrtPrice string `json:"sqrtPrice"`
	Tick      string `json:"tick"`
	Token0    Token  `json:"token0"`
	Token1    Token  `json:"token1"`
}

type Tick struct {
	TickIdx string `json:"tickIdx"`
}

type Position struct {
	ID        string `json:"id"`
	Liquidity string `json:"liquidity"`
	Pool      Pool   `json:"pool"`
	TickLower Tick   `json:"tickLower"`
	TickUpper Tick   `json:"tickUpper"`
}

type PositionsResponse struct {
	Data struct {
		Positions []Position `json:"positions"`
	} `json:"data"`
}

func (c *Client) GetPositions(ids []string) ([]Position, error) {
	idsStr := strings.Join(ids, ",")
	query := fmt.Sprintf("{\n  positions(where: {id_in: [%s]}) {\n    id\n    liquidity\n    pool {\n      sqrtPrice\n      tick\n      token0 {\n        symbol\n        decimals\n      }\n      token1 {\n        symbol\n        decimals\n      }\n    }\n    tickLower {\n      tickIdx\n    }\n    tickUpper {\n      tickIdx\n    }\n  }\n}", idsStr)
	req := map[string]string{
		"query": query,
	}

	resp, err := c.Post(c.baseURL, req)
	if err != nil {
		return nil, err
	}

	var posResp PositionsResponse
	err = json.NewDecoder(resp.Body).Decode(&posResp)
	if err != nil {
		return nil, err
	}

	return posResp.Data.Positions, nil
}
