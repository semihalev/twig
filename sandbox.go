package twig

import (
	"fmt"
)

// SecurityPolicy defines what's allowed in a sandboxed template context
type SecurityPolicy interface {
	// Function permissions
	IsFunctionAllowed(function string) bool

	// Filter permissions
	IsFilterAllowed(filter string) bool

	// Tag permissions
	IsTagAllowed(tag string) bool
}

// DefaultSecurityPolicy implements a simple security policy
type DefaultSecurityPolicy struct {
	AllowedFunctions map[string]bool
	AllowedFilters   map[string]bool
	AllowedTags      map[string]bool
}

// NewDefaultSecurityPolicy creates a security policy with safe defaults
func NewDefaultSecurityPolicy() *DefaultSecurityPolicy {
	return &DefaultSecurityPolicy{
		AllowedFunctions: map[string]bool{
			// Basic functions
			"range":  true,
			"cycle":  true,
			"date":   true,
			"min":    true,
			"max":    true,
			"random": true,
			"length": true,
			"merge":  true,
		},
		AllowedFilters: map[string]bool{
			// Basic filters
			"escape":     true,
			"e":          true,
			"raw":        true,
			"length":     true,
			"count":      true,
			"lower":      true,
			"upper":      true,
			"title":      true,
			"capitalize": true,
			"trim":       true,
			"nl2br":      true,
			"join":       true,
			"split":      true,
			"default":    true,
			"date":       true,
			"abs":        true,
			"first":      true,
			"last":       true,
			"reverse":    true,
			"sort":       true,
			"slice":      true,
		},
		AllowedTags: map[string]bool{
			// Basic control tags
			"if":       true,
			"else":     true,
			"elseif":   true,
			"for":      true,
			"set":      true,
			"verbatim": true,
		},
	}
}

// IsFunctionAllowed checks if a function is allowed
func (p *DefaultSecurityPolicy) IsFunctionAllowed(function string) bool {
	return p.AllowedFunctions[function]
}

// IsFilterAllowed checks if a filter is allowed
func (p *DefaultSecurityPolicy) IsFilterAllowed(filter string) bool {
	return p.AllowedFilters[filter]
}

// IsTagAllowed checks if a tag is allowed
func (p *DefaultSecurityPolicy) IsTagAllowed(tag string) bool {
	return p.AllowedTags[tag]
}

// SecurityViolation represents a sandbox security violation
type SecurityViolation struct {
	Message string
}

// Error returns the error message
func (v *SecurityViolation) Error() string {
	return fmt.Sprintf("Sandbox security violation: %s", v.Message)
}

// NewFunctionViolation creates a function security violation
func NewFunctionViolation(function string) error {
	return &SecurityViolation{
		Message: fmt.Sprintf("Function '%s' is not allowed in sandbox mode", function),
	}
}

// NewFilterViolation creates a filter security violation
func NewFilterViolation(filter string) error {
	return &SecurityViolation{
		Message: fmt.Sprintf("Filter '%s' is not allowed in sandbox mode", filter),
	}
}
