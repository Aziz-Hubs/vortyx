// Package errors provides a centralized error handling framework for the Vortyx backend.
// This package defines custom error types that follow consistent conventions across all services.
package errors

import (
	"errors"
	"fmt"
)

var (
	// Base error types
	ErrInternal     = errors.New("internal server error")
	ErrNotFound     = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrTimeout      = errors.New("operation timed out")
	ErrNotImplemented = errors.New("not implemented")
	ErrBadGateway   = errors.New("bad gateway")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// -----------------------------------------------------------------------------
// Error Code Type
// -----------------------------------------------------------------------------

// Code represents a machine-readable error code.
type Code string

const (
	CodeInternal          Code = "INTERNAL"
	CodeNotFound         Code = "NOT_FOUND"
	CodeAlreadyExists    Code = "ALREADY_EXISTS"
	CodeInvalidInput     Code = "INVALID_INPUT"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeForbidden        Code = "FORBIDDEN"
	CodeConflict         Code = "CONFLICT"
	CodeTimeout          Code = "TIMEOUT"
	CodeNotImplemented   Code = "NOT_IMPLEMENTED"
	CodeBadGateway       Code = "BAD_GATEWAY"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
)

// -----------------------------------------------------------------------------
// VortyxError
// -----------------------------------------------------------------------------

// VortyxError is the standard error type for the Vortyx backend.
// It includes a code, message, details, and underlying error.
type VortyxError struct {
	Code      Code   // Machine-readable error code
	Message   string // Human-readable error message
	Details   string // Additional details for debugging
	Err       error  // Underlying error
}

// Error implements the error interface.
func (e *VortyxError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is and errors.As support.
func (e *VortyxError) Unwrap() error {
	return e.Err
}

// WithDetails adds additional details to the error.
func (e *VortyxError) WithDetails(details string) *VortyxError {
	e.Details = details
	return e
}

// WithCause sets the underlying cause error.
func (e *VortyxError) WithCause(err error) *VortyxError {
	e.Err = err
	return e
}

// -----------------------------------------------------------------------------
// Constructor Functions
// -----------------------------------------------------------------------------

// New creates a new VortyxError with the given code and message.
func New(code Code, message string) *VortyxError {
	return &VortyxError{
		Code:    code,
		Message: message,
	}
}

// Internal creates an internal server error.
func Internal(message string) *VortyxError {
	return New(CodeInternal, message)
}

// NotFound creates a not found error.
func NotFound(message string) *VortyxError {
	return New(CodeNotFound, message)
}

// AlreadyExists creates an already exists error.
func AlreadyExists(message string) *VortyxError {
	return New(CodeAlreadyExists, message)
}

// InvalidInput creates an invalid input error.
func InvalidInput(message string) *VortyxError {
	return New(CodeInvalidInput, message)
}

// Unauthorized creates an unauthorized error.
func Unauthorized(message string) *VortyxError {
	return New(CodeUnauthorized, message)
}

// Forbidden creates a forbidden error.
func Forbidden(message string) *VortyxError {
	return New(CodeForbidden, message)
}

// Conflict creates a conflict error.
func Conflict(message string) *VortyxError {
	return New(CodeConflict, message)
}

// Timeout creates a timeout error.
func Timeout(message string) *VortyxError {
	return New(CodeTimeout, message)
}

// NotImplemented creates a not implemented error.
func NotImplemented(message string) *VortyxError {
	return New(CodeNotImplemented, message)
}

// BadGateway creates a bad gateway error.
func BadGateway(message string) *VortyxError {
	return New(CodeBadGateway, message)
}

// ServiceUnavailable creates a service unavailable error.
func ServiceUnavailable(message string) *VortyxError {
	return New(CodeServiceUnavailable, message)
}

// -----------------------------------------------------------------------------
// Wrap Functions
// -----------------------------------------------------------------------------

// Wrap wraps an existing error with a VortyxError.
func Wrap(code Code, message string, err error) *VortyxError {
	return New(code, message).WithCause(err)
}

// WrapInternal wraps an internal error.
func WrapInternal(message string, err error) *VortyxError {
	return Internal(message).WithCause(err)
}

// WrapNotFound wraps a not found error.
func WrapNotFound(message string, err error) *VortyxError {
	return NotFound(message).WithCause(err)
}

// WrapInvalidInput wraps an invalid input error.
func WrapInvalidInput(message string, err error) *VortyxError {
	return InvalidInput(message).WithCause(err)
}

// -----------------------------------------------------------------------------
// Validation Errors
// -----------------------------------------------------------------------------

// ValidationError represents a validation error for a specific field.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	msg := "validation failed:"
	for _, err := range e {
		msg += fmt.Sprintf(" %s", err.Error())
	}
	return msg
}

// Add adds a validation error to the collection.
func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// -----------------------------------------------------------------------------
// HTTP Status Mapping
// -----------------------------------------------------------------------------

// ToHTTPStatus maps an error code to an HTTP status code.
func (e *VortyxError) ToHTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return 404
	case CodeAlreadyExists:
		return 409
	case CodeInvalidInput:
		return 400
	case CodeUnauthorized:
		return 401
	case CodeForbidden:
		return 403
	case CodeConflict:
		return 409
	case CodeTimeout:
		return 504
	case CodeNotImplemented:
		return 501
	case CodeBadGateway:
		return 502
	case CodeServiceUnavailable:
		return 503
	default:
		return 500
	}
}

// ToConnectCode maps an error code to a ConnectRPC error code.
func (e *VortyxError) ToConnectCode() string {
	switch e.Code {
	case CodeNotFound:
		return "NOT_FOUND"
	case CodeAlreadyExists:
		return "ALREADY_EXISTS"
	case CodeInvalidInput:
		return "INVALID_ARGUMENT"
	case CodeUnauthorized:
		return "UNAUTHENTICATED"
	case CodeForbidden:
		return "PERMISSION_DENIED"
	case CodeConflict:
		return "ABORTED"
	case CodeTimeout:
		return "DEADLINE_EXCEEDED"
	case CodeNotImplemented:
		return "UNIMPLEMENTED"
	case CodeBadGateway:
		return "BAD_GATEWAY"
	case CodeServiceUnavailable:
		return "UNAVAILABLE"
	default:
		return "INTERNAL"
	}
}
