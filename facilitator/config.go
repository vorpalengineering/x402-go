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
	PrivateKey  string                    `yaml:"-"`
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
	var facilitatorConfig FacilitatorConfig
	if err := yaml.Unmarshal(data, &facilitatorConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Load secrets from environment variables
	if err := loadEnvVars(&facilitatorConfig); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	// TODO: validate config

	return &facilitatorConfig, nil
}

func (config *FacilitatorConfig) GetNetworkConfig(network string) (NetworkConfig, error) {
	networkConfig, exists := config.Networks[network]
	if !exists {
		return NetworkConfig{}, fmt.Errorf("network not configured: %s", network)
	}
	return networkConfig, nil
}

func (config *FacilitatorConfig) IsSupported(scheme, network string) bool {
	for _, s := range config.Supported {
		if s.Scheme == scheme && s.Network == network {
			return true
		}
	}
	return false
}

func loadEnvVars(config *FacilitatorConfig) error {
	// Load from environment variable
	// ex: export X402_FACILITATOR_PRIVATE_KEY=0x123...
	privateKey := os.Getenv("X402_FACILITATOR_PRIVATE_KEY")
	if privateKey == "" {
		return fmt.Errorf("X402_FACILITATOR_PRIVATE_KEY environment variable required")
	}
	config.PrivateKey = privateKey

	return nil
}
