package types

// Client/Facilitator types

type VerifyRequest struct {
	X402Version         int                 `json:"x402Version"`
	PaymentHeader       string              `json:"paymentHeader"` // Raw base64 encoded header
	PaymentRequirements PaymentRequirements `json:"paymentRequirements"`
}

type VerifyResponse struct {
	IsValid       bool   `json:"isValid"`
	InvalidReason string `json:"invalidReason,omitempty"`
}

type SettleRequest struct {
	X402Version         int                 `json:"x402Version"`
	PaymentHeader       string              `json:"paymentHeader"` // Raw base64 encoded header
	PaymentRequirements PaymentRequirements `json:"paymentRequirements"`
}

type SettleResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	TxHash    string `json:"txHash,omitempty"`
	NetworkId string `json:"networkId,omitempty"`
}

type SchemeNetwork struct {
	Scheme  string `json:"scheme"`
	Network string `json:"network"`
}

type SupportedResponse struct {
	Kinds []SchemeNetwork `json:"kinds"`
}

// Payment types

type PaymentRequiredResponse struct {
	X402Version int                   `json:"x402Version"`
	Accepts     []PaymentRequirements `json:"accepts"`
	Error       string                `json:"error,omitempty"`
}

type PaymentRequirements struct {
	Scheme            string         `json:"scheme"`
	Network           string         `json:"network"`
	MaxAmountRequired string         `json:"maxAmountRequired"`
	Resource          string         `json:"resource"`
	Description       string         `json:"description"`
	MimeType          string         `json:"mimeType"`
	OutputSchema      map[string]any `json:"outputSchema,omitempty"`
	PayTo             string         `json:"payTo"`
	MaxTimeoutSeconds int            `json:"maxTimeoutSeconds"`
	Asset             string         `json:"asset"`
	Extra             map[string]any `json:"extra,omitempty"`
}

type PaymentPayload struct {
	X402Version int            `json:"x402Version"`
	Scheme      string         `json:"scheme"`
	Network     string         `json:"network"`
	Payload     map[string]any `json:"payload"`
}
