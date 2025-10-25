package facilitator

import (
	"fmt"
	"os"

	"github.com/vorpalengineering/x402-go/types"
	"gopkg.in/yaml.v3"
)

type FacilitatorConfig struct {
	Server      ServerConfig              `yaml:"server"`
	Networks    map[string]NetworkConfig  `yaml:"networks"`
	Supported   []types.SchemeNetworkPair `yaml:"supported"`
	Transaction TransactionConfig         `yaml:"transaction"`
	Log         LogConfig                 `yaml:"log"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type NetworkConfig struct {
	RpcUrl  string `yaml:"rpc_url"`
	ChainId string `yaml:"chain_id"`
}

type TransactionConfig struct {
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	MaxGasPrice    string `yaml:"max_gas_price"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func LoadConfig(configPath string) (*FacilitatorConfig, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var conf FacilitatorConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// TODO: get secrets

	// TODO: validate config

	return &conf, nil
}
