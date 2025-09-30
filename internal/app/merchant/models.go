package merchant

import (
	"errors"
	// "time"
)

type Location struct {
	Latitude  float64 `json:"lat" validate:"required,latitude"`
	Longitude float64 `json:"long" validate:"required,longitude"`
}

type PostMerchantRequest struct {
	Name             string   `json:"name" validate:"required,min=2,max=30"`
	MerchantCategory string   `json:"merchantCategory" validate:"required,merchantCategory"`
	ImageURL         string   `json:"imageUrl" validate:"required,url,urlSuffix"`
	Location         Location `json:"location" validate:"required"`
}

type PostMerchantResponse struct {
	MerchantID string `json:"merchantId"`
}

// MerchantFilter holds filter params for searching merchants
type MerchantFilter struct {
	MerchantID       string
	Name             string
	MerchantCategory string
	CreatedAtSort    string
	Offset           int
	Limit            int
}

type GetMerchantsResponse struct {
	Data []Merchant `json:"data"`
	Meta Meta           `json:"meta"`
}

type Merchant struct {
	MerchantID       string   `json:"merchantId"`
	Name             string   `json:"name"`
	MerchantCategory string   `json:"merchantCategory"`
	ImageURL         string   `json:"imageUrl"`
	Location         Location `json:"location"`
	CreatedAt        string   `json:"createdAt"`
}

type Meta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// Domain errors for user operations
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// ErrorResponse represents the structure for error responses
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidationError represents validation errors with field-specific details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrorResponse represents the response for validation errors
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err string, message string) ErrorResponse {
	return ErrorResponse{
		Error:   err,
		Message: message,
	}
}

// NewErrorResponseWithDetails creates a new error response with details
func NewErrorResponseWithDetails(err string, message string, details map[string]string) ErrorResponse {
	return ErrorResponse{
		Error:   err,
		Message: message,
		Details: details,
	}
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(message string, errors []ValidationError) ValidationErrorResponse {
	return ValidationErrorResponse{
		Error:   "validation error",
		Message: message,
		Errors:  errors,
	}
}
