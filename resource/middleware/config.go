package middleware

import (
	"errors"

	"github.com/vorpalengineering/x402-go/types"
)

type MiddlewareConfig struct {
	// FacilitatorURL is the base URL of the x402 facilitator service
	FacilitatorURL string

	// DefaultRequirements specifies the default payment requirements
	// for protected routes that don't have specific requirements
	DefaultRequirements types.PaymentRequirements

	// ProtectedPaths is a list of path patterns that require payment
	// Supports glob patterns like "/api/*" or exact paths like "/data"
	ProtectedPaths []string

	// RouteRequirements maps specific routes to custom payment requirements
	// If a route matches multiple patterns, the most specific match is used
	// Routes not in this map will use DefaultRequirements
	RouteRequirements map[string]types.PaymentRequirements

	// RouteResources maps a specific route to its ResourceInfo
	RouteResources map[string]*types.ResourceInfo

	// PaymentHeaderName is the name of the HTTP header containing the payment signature
	// Defaults to "PAYMENT-SIGNATURE" if not specified
	PaymentHeaderName string
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
