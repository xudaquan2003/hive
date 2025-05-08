package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	OpReth struct {
		IP          string `yaml:"ip"`
		HTTPPort    uint16 `yaml:"http_port"`
		WSPort      uint16 `yaml:"ws_port"`
		AuthrpcPort uint16 `yaml:"authrpc_port"`
	} `yaml:"op-reth"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
