package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Client/Facilitator types

type VerifyRequest struct {
	PaymentPayload      PaymentPayload      `json:"paymentPayload"`
	PaymentRequirements PaymentRequirements `json:"paymentRequirements"`
}

type VerifyResponse struct {
	IsValid       bool   `json:"isValid"`
	InvalidReason string `json:"invalidReason,omitempty"`
	Payer         string `json:"payer,omitempty"`
}

type SettleRequest struct {
	PaymentPayload      PaymentPayload      `json:"paymentPayload"`
	PaymentRequirements PaymentRequirements `json:"paymentRequirements"`
}

type SettleResponse struct {
	Success     bool   `json:"success"`
	ErrorReason string `json:"errorReason,omitempty"`
	Payer       string `json:"payer,omitempty"`
	Transaction string `json:"transaction"`
	Network     string `json:"network"`
}

type SupportedKind struct {
	X402Version int            `json:"x402Version"`
	Scheme      string         `json:"scheme" yaml:"scheme"`
	Network     string         `json:"network" yaml:"network"`
	Extra       map[string]any `json:"extra,omitempty"`
}

type SupportedResponse struct {
	Kinds      []SupportedKind     `json:"kinds"`
	Extensions []string            `json:"extensions"`
	Signers    map[string][]string `json:"signers"`
}

// Payment types

type PaymentRequired struct {
	X402Version int                   `json:"x402Version"`
	Error       string                `json:"error,omitempty"`
	Resource    *ResourceInfo         `json:"resource,omitempty"`
	Accepts     []PaymentRequirements `json:"accepts"`
	Extensions  map[string]Extension  `json:"extensions,omitempty"`
}

type PaymentRequirements struct {
	Scheme            string         `json:"scheme" yaml:"scheme"`
	Network           string         `json:"network" yaml:"network"`
	Amount            string         `json:"amount" yaml:"amount"`
	Asset             string         `json:"asset" yaml:"asset"`
	PayTo             string         `json:"payTo" yaml:"pay_to"`
	MaxTimeoutSeconds int            `json:"maxTimeoutSeconds" yaml:"max_timeout_seconds"`
	Extra             map[string]any `json:"extra,omitempty" yaml:"extra,omitempty"`
}

type ResourceInfo struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type Extension struct {
	Info   map[string]any `json:"info"`
	Schema map[string]any `json:"schema"`
}

type PaymentPayload struct {
	X402Version int                 `json:"x402Version"`
	Resource    *ResourceInfo       `json:"resource,omitempty"`
	Accepted    PaymentRequirements `json:"accepted"`
	Payload     map[string]any      `json:"payload"`
	Extensions  map[string]any      `json:"extensions,omitempty"`
}

type ExactEVMSchemePayload struct {
	Signature     string                      `json:"signature"`
	Authorization ExactEVMSchemeAuthorization `json:"authorization"`
}

type ExactEVMSchemeAuthorization struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	ValidAfter  int64  `json:"validAfter"`
	ValidBefore int64  `json:"validBefore"`
	Nonce       string `json:"nonce"`
}

type EIP3009Authorization struct {
	From        common.Address
	To          common.Address
	Value       *big.Int
	ValidAfter  *big.Int
	ValidBefore *big.Int
	Nonce       [32]byte
	V           uint8
	R           [32]byte
	S           [32]byte
}
