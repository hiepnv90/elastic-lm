package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type TokenInstrument struct {
	Token      string `yaml:"token"`
	Instrument string `yaml:"instrument"`
}

type Binance struct {
	APIKey    string            `yaml:"api_key"`
	SecretKey string            `yaml:"secret_key"`
	Symbols   []TokenInstrument `yaml:"symbols"`
}

type SQLite struct {
	DBName string `yaml:"db_name"`
	Reset  bool   `yaml:"reset"`
}

type Config struct {
	Debug              bool     `yaml:"debug"`
	GraphQL            string   `yaml:"graphql"`
	Positions          []string `yaml:"positions"`
	Binance            Binance  `yaml:"binance"`
	AmountThresholdBps int      `yaml:"amount_threshold_bps"`
	SQLite             SQLite   `yaml:"sqlite"`
}

func Default() *Config {
	return &Config{
		Debug:     false,
		GraphQL:   "https://api.thegraph.com/subgraphs/name/kybernetwork/kyberswap-elastic-matic",
		Positions: []string{},
		Binance: Binance{
			APIKey:    "",
			SecretKey: "",
			Symbols:   nil,
		},
		AmountThresholdBps: 0,
		SQLite: SQLite{
			DBName: "elastic-lm.db",
		},
	}
}

func FromFile(fpath string) (*Config, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	cfg := Default()
	err = yaml.Unmarshal(data, cfg)
	return cfg, err
}
