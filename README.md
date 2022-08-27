# Elastic Liquidity Mining Helper

This is a very simple program that helps monitoring farming positions on kyberswap.com and open/adjust corresponding short positions on Binance for hedging.

## Quick Start

Clone code:
```bash
git clone https://github.com/hiepnv90/elastic-lm.git
cd elastic-lm
```

Example config:
```yaml
debug: false
graphql: "https://api.thegraph.com/subgraphs/name/kybernetwork/kyberswap-elastic-matic" # URL endpoint to thegraph's GraphQL
positions: ["0", "1"] # your farming positions on kyberswap.com
binance:
  api_key: "test_binance_api_key"
  secret_key: "test_secret_key"
amount_threshold_bps: 100 # 1% of max amount0 at min price (lower tick)
```

Run program:
```bash
go run ./cmd/elastic-lm/main.go --config elastic-lm.yaml
```

Build program:
```bash
go build ./cmd/elastic-lm
```

## Limitations
1. The program don't store Binance's positions on persistent storage, so the information will be reseted when the program restarted.
1. The program opens Binance's short positions using market orders it not suitable for opening big positions.
1. This is a very simple program, users need to take some actions for it to work.

## New Updates
- Support for adjusting Binance's positions with an amount threshold in BPS.
