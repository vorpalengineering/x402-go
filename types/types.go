package types

type VerifyRequest struct {
	PaymentPayload interface{} `json:"payment_payload"`
	Requirements   interface{} `json:"requirements"`
}

type VerifyResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

type SettleRequest struct {
	PaymentPayload interface{} `json:"payment_payload"`
}

type SettleResponse struct {
	TxHash  string `json:"tx_hash,omitempty"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
