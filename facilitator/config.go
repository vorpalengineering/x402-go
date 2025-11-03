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

	// Validate config
	if err := facilitatorConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

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

func (config *FacilitatorConfig) Validate() error {
	// Validate server config
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be 1-65535)", config.Server.Port)
	}

	// Validate networks
	if len(config.Networks) == 0 {
		return fmt.Errorf("at least one network must be configured")
	}

	for network, netCfg := range config.Networks {
		if netCfg.RpcUrl == "" {
			return fmt.Errorf("network %s missing rpc_url", network)
		}
		if netCfg.ChainId == "" {
			return fmt.Errorf("network %s missing chain_id", network)
		}
	}

	// Validate supported schemes reference valid networks
	for _, pair := range config.Supported {
		if pair.Scheme == "" {
			return fmt.Errorf("supported scheme cannot be empty")
		}
		if pair.Network == "" {
			return fmt.Errorf("supported network cannot be empty")
		}
		if _, exists := config.Networks[pair.Network]; !exists {
			return fmt.Errorf("supported network %s is not defined in networks config", pair.Network)
		}
	}

	// Validate transaction config
	if config.Transaction.TimeoutSeconds <= 0 {
		return fmt.Errorf("transaction timeout must be positive, got %d", config.Transaction.TimeoutSeconds)
	}
	if config.Transaction.MaxGasPrice == "" {
		return fmt.Errorf("transaction max_gas_price must be set")
	}

	// Validate log config
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[config.Log.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", config.Log.Level)
	}

	// Validate private key is set
	if config.PrivateKey == "" {
		return fmt.Errorf("private key must be set")
	}

	return nil
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
