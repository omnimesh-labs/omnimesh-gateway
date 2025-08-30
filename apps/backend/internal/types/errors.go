package types

import (
	"fmt"
	"net/http"
)

// Error represents a structured error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Status  int    `json:"-"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   *Error `json:"error"`
	Success bool   `json:"success"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface for JSONRPCError
func (e *JSONRPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("JSON-RPC error %d: %s (%v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}


// Predefined error codes
const (
	// Authentication errors
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       = "TOKEN_INVALID"
	ErrCodeInsufficientRights = "INSUFFICIENT_RIGHTS"

	// Validation errors
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeMissingRequired  = "MISSING_REQUIRED_FIELD"

	// Resource errors
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
	ErrCodeConflict      = "CONFLICT"

	// Rate limiting errors
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeQuotaExceeded     = "QUOTA_EXCEEDED"

	// Server errors
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            = "TIMEOUT"
	ErrCodeBadGateway         = "BAD_GATEWAY"

	// Gateway errors
	ErrCodeServerNotFound     = "SERVER_NOT_FOUND"
	ErrCodeServerUnhealthy    = "SERVER_UNHEALTHY"
	ErrCodeProxyError         = "PROXY_ERROR"
	ErrCodeCircuitBreakerOpen = "CIRCUIT_BREAKER_OPEN"

	// Policy errors
	ErrCodePolicyViolation = "POLICY_VIOLATION"
	ErrCodeAccessDenied    = "ACCESS_DENIED"
)

// NewError creates a new structured error
func NewError(code, message string, status int) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// NewErrorWithDetails creates a new structured error with details
func NewErrorWithDetails(code, message, details string, status int) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: details,
		Status:  status,
	}
}

// Authentication error constructors
func NewUnauthorizedError(message string) *Error {
	return NewError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func NewInvalidCredentialsError() *Error {
	return NewError(ErrCodeInvalidCredentials, "Invalid email or password", http.StatusUnauthorized)
}

func NewTokenExpiredError() *Error {
	return NewError(ErrCodeTokenExpired, "Token has expired", http.StatusUnauthorized)
}

func NewTokenInvalidError() *Error {
	return NewError(ErrCodeTokenInvalid, "Invalid token", http.StatusUnauthorized)
}

func NewInsufficientRightsError() *Error {
	return NewError(ErrCodeInsufficientRights, "Insufficient rights to perform this action", http.StatusForbidden)
}

func NewForbiddenError(message string) *Error {
	return NewError(ErrCodeAccessDenied, message, http.StatusForbidden)
}

// Validation error constructors
func NewValidationError(message string) *Error {
	return NewError(ErrCodeValidationFailed, message, http.StatusBadRequest)
}

func NewInvalidInputError(message string) *Error {
	return NewError(ErrCodeInvalidInput, message, http.StatusBadRequest)
}

func NewMissingRequiredFieldError(field string) *Error {
	return NewError(ErrCodeMissingRequired, fmt.Sprintf("Missing required field: %s", field), http.StatusBadRequest)
}

// Resource error constructors
func NewNotFoundError(message string) *Error {
	return NewError(ErrCodeNotFound, message, http.StatusNotFound)
}

func NewAlreadyExistsError(message string) *Error {
	return NewError(ErrCodeAlreadyExists, message, http.StatusConflict)
}

func NewConflictError(message string) *Error {
	return NewError(ErrCodeConflict, message, http.StatusConflict)
}

// Rate limiting error constructors
func NewRateLimitExceededError(message string) *Error {
	return NewError(ErrCodeRateLimitExceeded, message, http.StatusTooManyRequests)
}

func NewQuotaExceededError(message string) *Error {
	return NewError(ErrCodeQuotaExceeded, message, http.StatusTooManyRequests)
}

// Server error constructors
func NewInternalError(message string) *Error {
	return NewError(ErrCodeInternalError, message, http.StatusInternalServerError)
}

func NewServiceUnavailableError(message string) *Error {
	return NewError(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

func NewNotImplementedError(message string) *Error {
	return NewError("NOT_IMPLEMENTED", message, http.StatusNotImplemented)
}

func NewTimeoutError(message string) *Error {
	return NewError(ErrCodeTimeout, message, http.StatusGatewayTimeout)
}

func NewBadGatewayError(message string) *Error {
	return NewError(ErrCodeBadGateway, message, http.StatusBadGateway)
}

// Gateway error constructors
func NewServerNotFoundError(serverID string) *Error {
	return NewError(ErrCodeServerNotFound, fmt.Sprintf("MCP server not found: %s", serverID), http.StatusNotFound)
}

func NewServerUnhealthyError(serverID string) *Error {
	return NewError(ErrCodeServerUnhealthy, fmt.Sprintf("MCP server is unhealthy: %s", serverID), http.StatusServiceUnavailable)
}

func NewProxyError(message string) *Error {
	return NewError(ErrCodeProxyError, message, http.StatusBadGateway)
}

func NewCircuitBreakerOpenError(serverID string) *Error {
	return NewError(ErrCodeCircuitBreakerOpen, fmt.Sprintf("Circuit breaker is open for server: %s", serverID), http.StatusServiceUnavailable)
}

// Policy error constructors
func NewPolicyViolationError(policy string) *Error {
	return NewError(ErrCodePolicyViolation, fmt.Sprintf("Policy violation: %s", policy), http.StatusForbidden)
}

func NewAccessDeniedError(message string) *Error {
	return NewError(ErrCodeAccessDenied, message, http.StatusForbidden)
}

// IsError checks if an error is of a specific type
func IsError(err error, code string) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// GetStatusCode returns the HTTP status code for an error
func GetStatusCode(err error) int {
	if e, ok := err.(*Error); ok {
		return e.Status
	}
	return http.StatusInternalServerError
}
