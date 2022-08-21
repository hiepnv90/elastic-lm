package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Debug    bool   `yaml:"debug"`
	GraphQL  string `yaml:"graphql"`
	Position string `yaml:"position"`
}

func Default() *Config {
	return &Config{
		Debug:    true,
		GraphQL:  "https://api.thegraph.com/subgraphs/name/kybernetwork/kyberswap-elastic-matic",
		Position: "0",
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
