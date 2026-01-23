package facilitator

import (
	"testing"

	"github.com/vorpalengineering/x402-go/types"
)

func TestValidateConfig(t *testing.T) {
	validConfig := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"base": {
				RpcUrl:  "https://mainnet.base.org",
				ChainId: "8453",
			},
		},
		Supported: []types.SupportedKind{
			{Scheme: "exact", Network: "base"},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x1234567890abcdef",
	}

	err := validConfig.Validate()
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}
}

func TestValidateInvalidPort(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 0, // Invalid
		},
		Networks: map[string]NetworkConfig{
			"base": {RpcUrl: "https://mainnet.base.org", ChainId: "8453"},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x1234",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for invalid port, got nil")
	}
}

func TestValidateNoNetworks(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{}, // Empty
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x1234",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for no networks configured, got nil")
	}
}

func TestValidateMissingRpcUrl(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"base": {
				RpcUrl:  "", // Missing
				ChainId: "8453",
			},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x1234",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for missing rpc_url, got nil")
	}
}

func TestValidateMissingChainId(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"base": {
				RpcUrl:  "https://mainnet.base.org",
				ChainId: "", // Missing
			},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x1234",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for missing chain_id, got nil")
	}
}

func TestValidateUndefinedNetwork(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"base": {
				RpcUrl:  "https://mainnet.base.org",
				ChainId: "8453",
			},
		},
		Supported: []types.SupportedKind{
			{Scheme: "exact", Network: "ethereum"}, // Network not defined
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x1234",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for undefined network in supported schemes, got nil")
	}
}

func TestValidateEmptyScheme(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"base": {
				RpcUrl:  "https://mainnet.base.org",
				ChainId: "8453",
			},
		},
		Supported: []types.SupportedKind{
			{Scheme: "", Network: "base"}, // Empty scheme
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x1234",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for empty scheme, got nil")
	}
}

func TestValidateInvalidTimeout(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"base": {RpcUrl: "https://mainnet.base.org", ChainId: "8453"},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 0, // Invalid
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x1234",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for invalid timeout, got nil")
	}
}

func TestValidateMissingMaxGasPrice(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"base": {RpcUrl: "https://mainnet.base.org", ChainId: "8453"},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "", // Missing
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x1234",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for missing max_gas_price, got nil")
	}
}

func TestValidateInvalidLogLevel(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"base": {RpcUrl: "https://mainnet.base.org", ChainId: "8453"},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "invalid", // Invalid
		},
		PrivateKey: "0x1234",
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for invalid log level, got nil")
	}
}

func TestValidateMissingPrivateKey(t *testing.T) {
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"base": {RpcUrl: "https://mainnet.base.org", ChainId: "8453"},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "", // Missing
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for missing private key, got nil")
	}
}
