package twig

import (
	"fmt"
	"reflect"
	"strings"
)

// SecurityPolicy defines what's allowed in a sandboxed template context
type SecurityPolicy interface {
	// Tag permissions
	IsTagAllowed(tag string) bool

	// Filter permissions
	IsFilterAllowed(filter string) bool

	// Function permissions
	IsFunctionAllowed(function string) bool

	// Property access permissions
	IsPropertyAllowed(objType string, property string) bool

	// Method call permissions
	IsMethodAllowed(objType string, method string) bool
}

// DefaultSecurityPolicy implements a standard security policy
type DefaultSecurityPolicy struct {
	// Allow lists
	AllowedTags       map[string]bool
	AllowedFilters    map[string]bool
	AllowedFunctions  map[string]bool
	AllowedProperties map[string]map[string]bool // Type -> Property -> Allowed
	AllowedMethods    map[string]map[string]bool // Type -> Method -> Allowed
}

// NewDefaultSecurityPolicy creates a default security policy with safe defaults
func NewDefaultSecurityPolicy() *DefaultSecurityPolicy {
	return &DefaultSecurityPolicy{
		AllowedTags: map[string]bool{
			"if":       true,
			"else":     true,
			"elseif":   true,
			"for":      true,
			"set":      true,
			"verbatim": true,
			"do":       true,
			// More safe tags
		},
		AllowedFilters: map[string]bool{
			"escape":    true,
			"e":         true,
			"raw":       true,
			"length":    true,
			"count":     true,
			"lower":     true,
			"upper":     true,
			"title":     true,
			"capitalize": true,
			"trim":      true,
			"nl2br":     true,
			"join":      true,
			"split":     true,
			"default":   true,
			"date":      true,
			"number_format": true,
			"abs":       true,
			"first":     true,
			"last":      true,
			"reverse":   true,
			"sort":      true,
			"slice":     true,
			// More safe filters
		},
		AllowedFunctions: map[string]bool{
			"range":     true,
			"cycle":     true,
			"constant":  true,
			"date":      true,
			"min":       true,
			"max":       true,
			"random":    true,
			// More safe functions
		},
		AllowedProperties: make(map[string]map[string]bool),
		AllowedMethods:    make(map[string]map[string]bool),
	}
}

// IsTagAllowed checks if a tag is allowed
func (p *DefaultSecurityPolicy) IsTagAllowed(tag string) bool {
	return p.AllowedTags[tag]
}

// IsFilterAllowed checks if a filter is allowed
func (p *DefaultSecurityPolicy) IsFilterAllowed(filter string) bool {
	return p.AllowedFilters[filter]
}

// IsFunctionAllowed checks if a function is allowed
func (p *DefaultSecurityPolicy) IsFunctionAllowed(function string) bool {
	return p.AllowedFunctions[function]
}

// IsPropertyAllowed checks if property access is allowed for a type
func (p *DefaultSecurityPolicy) IsPropertyAllowed(objType string, property string) bool {
	props, ok := p.AllowedProperties[objType]
	if !ok {
		return false
	}
	return props[property]
}

// IsMethodAllowed checks if method call is allowed for a type
func (p *DefaultSecurityPolicy) IsMethodAllowed(objType string, method string) bool {
	methods, ok := p.AllowedMethods[objType]
	if !ok {
		return false
	}
	return methods[method]
}

// AllowObjectType adds all properties and methods of a type to the allowlist
func (p *DefaultSecurityPolicy) AllowObjectType(obj interface{}) {
	t := reflect.TypeOf(obj)
	typeName := t.String()
	
	// Allow properties
	if p.AllowedProperties[typeName] == nil {
		p.AllowedProperties[typeName] = make(map[string]bool)
	}
	
	// Allow methods
	if p.AllowedMethods[typeName] == nil {
		p.AllowedMethods[typeName] = make(map[string]bool)
	}
	
	// Add all methods
	for i := 0; i < t.NumMethod(); i++ {
		methodName := t.Method(i).Name
		p.AllowedMethods[typeName][methodName] = true
	}
	
	// If it's a struct, add all fields
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			// Only allow exported fields
			if field.IsExported() {
				p.AllowedProperties[typeName][field.Name] = true
			}
		}
	}
}

// GetTypeString returns a string representation of an object's type
func GetTypeString(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t == nil {
		return "nil"
	}
	return t.String()
}

// SecurityViolation represents a security policy violation
type SecurityViolation struct {
	Message string
	Tag     string
	Filter  string
	Obj     string
	Access  string
}

// Error returns the string representation of this error
func (v *SecurityViolation) Error() string {
	return fmt.Sprintf("Sandbox security violation: %s", v.Message)
}

// NewTagViolation creates a new tag security violation
func NewTagViolation(tag string) *SecurityViolation {
	return &SecurityViolation{
		Message: fmt.Sprintf("Tag '%s' is not allowed in sandbox mode", tag),
		Tag:     tag,
	}
}

// NewFilterViolation creates a new filter security violation
func NewFilterViolation(filter string) *SecurityViolation {
	return &SecurityViolation{
		Message: fmt.Sprintf("Filter '%s' is not allowed in sandbox mode", filter),
		Filter:  filter,
	}
}

// NewPropertyViolation creates a new property access security violation
func NewPropertyViolation(obj interface{}, property string) *SecurityViolation {
	objType := GetTypeString(obj)
	return &SecurityViolation{
		Message: fmt.Sprintf("Property '%s' of type '%s' is not allowed in sandbox mode", 
			property, objType),
		Obj:     objType,
		Access:  property,
	}
}

// NewMethodViolation creates a new method call security violation
func NewMethodViolation(obj interface{}, method string) *SecurityViolation {
	objType := GetTypeString(obj)
	return &SecurityViolation{
		Message: fmt.Sprintf("Method '%s' of type '%s' is not allowed in sandbox mode", 
			method, objType),
		Obj:     objType,
		Access:  method,
	}
}

// NewFunctionViolation creates a new function call security violation
func NewFunctionViolation(function string) *SecurityViolation {
	return &SecurityViolation{
		Message: fmt.Sprintf("Function '%s' is not allowed in sandbox mode", function),
		Access:  function,
	}
}