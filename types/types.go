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
	Success     bool   `json:"success"`
	ErrorReason string `json:"errorReason,omitempty"`
	Transaction string `json:"transaction,omitempty"`
	Network     string `json:"network,omitempty"`
	Payer       string `json:"payer,omitempty"`
}

type SchemeNetworkPair struct {
	Scheme  string `json:"scheme" yaml:"scheme"`
	Network string `json:"network" yaml:"network"`
}

type SupportedResponse struct {
	Kinds []SchemeNetworkPair `json:"kinds"`
}

// Payment types

type PaymentRequiredResponse struct {
	X402Version int                   `json:"x402Version"`
	Accepts     []PaymentRequirements `json:"accepts"`
	Error       string                `json:"error,omitempty"`
}

type PaymentRequirements struct {
	Scheme            string         `json:"scheme" yaml:"scheme"`
	Network           string         `json:"network" yaml:"network"`
	MaxAmountRequired string         `json:"maxAmountRequired" yaml:"max_amount_required"`
	Resource          string         `json:"resource" yaml:"resource"`
	Description       string         `json:"description" yaml:"description"`
	MimeType          string         `json:"mimeType" yaml:"mime_type"`
	OutputSchema      map[string]any `json:"outputSchema,omitempty" yaml:"output_scheme,omitempty"`
	PayTo             string         `json:"payTo" yaml:"pay_to"`
	MaxTimeoutSeconds int            `json:"maxTimeoutSeconds" yaml:"max_timeout_seconds"`
	Asset             string         `json:"asset" yaml:"asset"`
	Extra             map[string]any `json:"extra,omitempty" yaml:"extra,omitempty"`
}

type PaymentPayload struct {
	X402Version int            `json:"x402Version"`
	Scheme      string         `json:"scheme"`
	Network     string         `json:"network"`
	Payload     map[string]any `json:"payload"`
}

type ExactSchemePayload struct {
	Signature     string                   `json:"signature"`
	Authorization ExactSchemeAuthorization `json:"authorization"`
}

type ExactSchemeAuthorization struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	ValidAfter  int64  `json:"validAfter"`
	ValidBefore int64  `json:"validBefore"`
	Nonce       string `json:"nonce"`
}
