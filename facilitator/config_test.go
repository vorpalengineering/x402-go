package facilitator

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vorpalengineering/x402-go/types"
)

func TestValidateConfig(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	validConfig := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"eip155:8453": {
				RpcUrl: "https://mainnet.base.org",
			},
		},
		Supported: []types.SupportedKind{
			{Scheme: "exact", Network: "eip155:8453"},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: privKey,
		},
	}

	err = validConfig.Validate()
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}
}

func TestValidateInvalidPort(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 0, // Invalid
		},
		Networks: map[string]NetworkConfig{
			"eip155:8453": {
				RpcUrl: "https://mainnet.base.org",
			},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: privKey,
		},
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected error for invalid port, got nil")
	}
}

func TestValidateNoNetworks(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
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
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: privKey,
		},
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected error for no networks configured, got nil")
	}
}

func TestValidateMissingRpcUrl(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"eip155:8453": {
				RpcUrl: "", // Missing
			},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: privKey,
		},
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected error for missing rpc_url, got nil")
	}
}

func TestValidateUndefinedNetwork(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"eip155:8453": {
				RpcUrl: "https://mainnet.base.org",
			},
		},
		Supported: []types.SupportedKind{
			{Scheme: "exact", Network: "eip155:1"}, // Network not defined
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: privKey,
		},
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected error for undefined network in supported schemes, got nil")
	}
}

func TestValidateEmptyScheme(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"eip155:8453": {
				RpcUrl: "https://mainnet.base.org",
			},
		},
		Supported: []types.SupportedKind{
			{Scheme: "", Network: "eip155:8453"}, // Empty scheme
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: privKey,
		},
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected error for empty scheme, got nil")
	}
}

func TestValidateInvalidTimeout(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"eip155:8453": {
				RpcUrl: "https://mainnet.base.org",
			},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 0, // Invalid
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: privKey,
		},
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected error for invalid timeout, got nil")
	}
}

func TestValidateMissingMaxGasPrice(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"eip155:8453": {
				RpcUrl: "https://mainnet.base.org",
			},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "", // Missing
		},
		Log: LogConfig{
			Level: "info",
		},
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: privKey,
		},
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected error for missing max_gas_price, got nil")
	}
}

func TestValidateInvalidLogLevel(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"eip155:8453": {
				RpcUrl: "https://mainnet.base.org",
			},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "invalid", // Invalid
		},
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: privKey,
		},
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected error for invalid log level, got nil")
	}
}

func TestValidateMissingPrivateKey(t *testing.T) {
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	config := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Networks: map[string]NetworkConfig{
			"eip155:8453": {
				RpcUrl: "https://mainnet.base.org",
			},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		Signer: SignerConfig{
			Address:    addr,
			PrivateKey: nil, // Missing
		},
	}

	err = config.Validate()
	if err == nil {
		t.Error("Expected error for missing private key, got nil")
	}
}
