// Package common provides shared validation utilities used across the Vortyx backend.
// This package contains common validation helpers to reduce code duplication.
package common

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUUID         = errors.New("invalid UUID format")
	ErrEmptyString         = errors.New("string cannot be empty")
	ErrTooLong             = errors.New("string exceeds maximum length")
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrInvalidIP           = errors.New("invalid IP address format")
	ErrInvalidURL          = errors.New("invalid URL format")
	ErrInvalidRange        = errors.New("value out of valid range")
	ErrInvalidFormat       = errors.New("invalid format")
	ErrTransactionStart    = errors.New("failed to start transaction")
	ErrTransactionCommit   = errors.New("failed to commit transaction")
	ErrTransactionRollback = errors.New("failed to rollback transaction")
)

// WrapError wraps an error with a context message.
func WrapError(base error, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", base, err)
}

// -----------------------------------------------------------------------------
// UUID Validation
// -----------------------------------------------------------------------------

// IsValidUUID checks if a string is a valid UUID.
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// ValidateUUID returns an error if the string is not a valid UUID.
func ValidateUUID(s string) error {
	if s == "" {
		return ErrEmptyString
	}
	if !IsValidUUID(s) {
		return ErrInvalidUUID
	}
	return nil
}

// ParseUUID parses a string and returns a UUID pointer, or an error.
func ParseUUID(s string) (*uuid.UUID, error) {
	if err := ValidateUUID(s); err != nil {
		return nil, err
	}
	id := uuid.MustParse(s)
	return &id, nil
}

// -----------------------------------------------------------------------------
// String Validation
// -----------------------------------------------------------------------------

// ValidateRequired checks if a string is non-empty after trimming whitespace.
func ValidateRequired(s string) error {
	if strings.TrimSpace(s) == "" {
		return ErrEmptyString
	}
	return nil
}

// ValidateLength checks if a string length is within the specified bounds.
func ValidateLength(s string, min, max int) error {
	if len(s) < min {
		return fmt.Errorf("%w: minimum length is %d", ErrInvalidRange, min)
	}
	if max > 0 && len(s) > max {
		return fmt.Errorf("%w: maximum length is %d", ErrTooLong, max)
	}
	return nil
}

// ValidatePattern checks if a string matches a regex pattern.
func ValidatePattern(s string, pattern string) error {
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidFormat, err)
	}
	if !matched {
		return ErrInvalidFormat
	}
	return nil
}

// -----------------------------------------------------------------------------
// Email Validation
// -----------------------------------------------------------------------------

// IsValidEmail checks if a string is a valid email address.
func IsValidEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

// ValidateEmail returns an error if the string is not a valid email.
func ValidateEmail(s string) error {
	if err := ValidateRequired(s); err != nil {
		return err
	}
	if !IsValidEmail(s) {
		return ErrInvalidEmail
	}
	return nil
}

// -----------------------------------------------------------------------------
// Network Validation
// -----------------------------------------------------------------------------

// IsValidIP checks if a string is a valid IP address (IPv4 or IPv6).
func IsValidIP(s string) bool {
	return net.ParseIP(s) != nil
}

// ValidateIP returns an error if the string is not a valid IP address.
func ValidateIP(s string) error {
	if err := ValidateRequired(s); err != nil {
		return err
	}
	if !IsValidIP(s) {
		return ErrInvalidIP
	}
	return nil
}

// ValidatePort checks if a port number is valid (1-65535).
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%w: port must be between 1 and 65535", ErrInvalidRange)
	}
	return nil
}

// -----------------------------------------------------------------------------
// URL Validation
// -----------------------------------------------------------------------------

var urlRegex = regexp.MustCompile(`^https?://`)

// IsValidURL checks if a string is a valid URL with http or https scheme.
func IsValidURL(s string) bool {
	if !urlRegex.MatchString(s) {
		return false
	}
	_, err := url.ParseRequestURI(s)
	return err == nil
}

// ValidateURL returns an error if the string is not a valid URL.
func ValidateURL(s string) error {
	if err := ValidateRequired(s); err != nil {
		return err
	}
	if !IsValidURL(s) {
		return ErrInvalidURL
	}
	return nil
}

// -----------------------------------------------------------------------------
// Time Validation
// -----------------------------------------------------------------------------

// IsValidTimeFormat checks if a string matches the ISO 8601 time format.
func IsValidTimeFormat(s string) bool {
	_, err := time.Parse(time.RFC3339, s)
	return err == nil
}

// ValidateTimeFormat returns an error if the string is not a valid ISO 8601 time.
func ValidateTimeFormat(s string) error {
	if err := ValidateRequired(s); err != nil {
		return err
	}
	if !IsValidTimeFormat(s) {
		return fmt.Errorf("%w: expected ISO 8601 format (e.g., 2024-01-15T10:30:00Z)", ErrInvalidFormat)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Range Validation
// -----------------------------------------------------------------------------

// ValidateIntRange checks if an integer is within the specified bounds.
func ValidateIntRange(i, min, max int) error {
	if i < min || i > max {
		return fmt.Errorf("%w: value %d must be between %d and %d", ErrInvalidRange, i, min, max)
	}
	return nil
}

// ValidateFloatRange checks if a float is within the specified bounds.
func ValidateFloatRange(f, min, max float64) error {
	if f < min || f > max {
		return fmt.Errorf("%w: value %f must be between %f and %f", ErrInvalidRange, f, min, max)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Slice Validation
// -----------------------------------------------------------------------------

// ValidateSliceNotEmpty checks if a slice is not empty.
func ValidateSliceNotEmpty[T any](s []T) error {
	if len(s) == 0 {
		return errors.New("slice cannot be empty")
	}
	return nil
}

// ValidateSliceUnique checks if a slice contains unique elements.
func ValidateSliceUnique[T comparable](s []T) error {
	seen := make(map[T]bool)
	for _, v := range s {
		if seen[v] {
			return fmt.Errorf("%w: duplicate value found", ErrInvalidFormat)
		}
		seen[v] = true
	}
	return nil
}

// -----------------------------------------------------------------------------
// Map Validation
// -----------------------------------------------------------------------------

// ValidateMapKeys checks if a map has the required keys.
func ValidateMapKeys(m map[string]interface{}, requiredKeys []string) error {
	for _, key := range requiredKeys {
		if _, ok := m[key]; !ok {
			return fmt.Errorf("missing required key: %s", key)
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Composite Validators
// -----------------------------------------------------------------------------

// ValidateAgentName validates an agent name.
func ValidateAgentName(name string) error {
	if err := ValidateRequired(name); err != nil {
		return err
	}
	if err := ValidateLength(name, 1, 255); err != nil {
		return err
	}
	return ValidatePattern(name, `^[a-zA-Z0-9][a-zA-Z0-9 _\-.]+$`)
}

// ValidateHostname validates a hostname.
func ValidateHostname(hostname string) error {
	if err := ValidateRequired(hostname); err != nil {
		return err
	}
	if err := ValidateLength(hostname, 1, 253); err != nil {
		return err
	}
	return ValidatePattern(hostname, `^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
}

// ValidateOSType validates an operating system type.
func ValidateOSType(osType string) error {
	validTypes := []string{"linux", "windows", "darwin", "freebsd", "netbsd", "openbsd"}
	osType = strings.ToLower(osType)
	for _, valid := range validTypes {
		if osType == valid {
			return nil
		}
	}
	return fmt.Errorf("%w: must be one of: %v", ErrInvalidFormat, validTypes)
}
