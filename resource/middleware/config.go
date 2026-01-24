package middleware

import (
	"errors"

	"github.com/vorpalengineering/x402-go/types"
)

type MiddlewareConfig struct {
	// FacilitatorURL is the base URL of the x402 facilitator service
	FacilitatorURL string `json:"facilitatorUrl" toml:"facilitator_url"`

	// DefaultRequirements specifies the default payment requirements
	// for protected routes that don't have specific requirements
	DefaultRequirements types.PaymentRequirements `json:"defaultRequirements" toml:"default_requirements"`

	// ProtectedPaths is a list of path patterns that require payment
	// Supports glob patterns like "/api/*" or exact paths like "/data"
	ProtectedPaths []string `json:"protectedPaths" toml:"protected_paths"`

	// RouteRequirements maps specific routes to custom payment requirements
	// If a route matches multiple patterns, the most specific match is used
	// Routes not in this map will use DefaultRequirements
	RouteRequirements map[string]types.PaymentRequirements `json:"routeRequirements,omitempty" toml:"route_requirements"`

	// RouteResources maps a specific route to its ResourceInfo
	RouteResources map[string]*types.ResourceInfo `json:"routeResources,omitempty" toml:"route_resources"`

	// PaymentHeaderName is the name of the HTTP header containing the payment signature
	// Defaults to "PAYMENT-SIGNATURE" if not specified
	PaymentHeaderName string `json:"paymentHeaderName,omitempty" toml:"payment_header_name"`

	// MaxBufferSize is the maximum response buffer size in bytes.
	// If the handler response exceeds this size, the request is aborted.
	// 0 means unlimited.
	MaxBufferSize int `json:"maxBufferSize,omitempty" toml:"max_buffer_size"`

	// DiscoveryEnabled enables serving the /.well-known/x402 discovery endpoint
	DiscoveryEnabled bool `json:"discoveryEnabled,omitempty" toml:"discovery_enabled"`

	// OwnershipProofs is a list of pre-generated EIP-191 signatures
	// proving ownership of the protected resource URLs
	OwnershipProofs []string `json:"ownershipProofs,omitempty" toml:"ownership_proofs"`

	// Instructions is an optional markdown-formatted string containing
	// instructions or information for users of your resources.
	// Included in the /.well-known/x402 discovery response if non-empty.
	Instructions string `json:"instructions,omitempty" toml:"instructions"`
}

func (c *MiddlewareConfig) Validate() error {
	// Check required variables
	if c.FacilitatorURL == "" {
		return errors.New("facilitator URL is required")
	}
	if len(c.ProtectedPaths) == 0 {
		return errors.New("at least one protected path must be specified")
	}

	// Validate default requirements
	if err := validatePaymentRequirements(&c.DefaultRequirements); err != nil {
		return errors.New("invalid default requirements: " + err.Error())
	}

	// Validate route-specific requirements
	for route, req := range c.RouteRequirements {
		if err := validatePaymentRequirements(&req); err != nil {
			return errors.New("invalid requirements for route " + route + ": " + err.Error())
		}
	}

	return nil
}

func (c *MiddlewareConfig) GetPaymentHeaderName() string {
	if c.PaymentHeaderName == "" {
		return "PAYMENT-SIGNATURE"
	}
	return c.PaymentHeaderName
}

func validatePaymentRequirements(req *types.PaymentRequirements) error {
	// Check required variables
	if req.Scheme == "" {
		return errors.New("scheme is required")
	}
	if req.Network == "" {
		return errors.New("network is required")
	}
	if req.Amount == "" {
		return errors.New("max amount required is required")
	}
	if req.PayTo == "" {
		return errors.New("pay to address is required")
	}
	if req.Asset == "" {
		return errors.New("asset address is required")
	}
	return nil
}
