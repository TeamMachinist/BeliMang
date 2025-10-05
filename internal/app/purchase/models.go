package purchase

import (
	"errors"
)

type UserLocation struct {
	Lat  float64 `json:"lat" validate:"required,gte=-90,lte=90"`
	Long float64 `json:"long" validate:"required,gte=-180,lte=180"`
}

type OrderItem struct {
	ItemID   string `json:"itemId" validate:"required"`
	Quantity int    `json:"quantity" validate:"required,min=1"`
}

type Order struct {
	MerchantID      string      `json:"merchantId" validate:"required"`
	IsStartingPoint bool        `json:"isStartingPoint"`
	Items           []OrderItem `json:"items" validate:"dive"`
}

type EstimateRequest struct {
	UserLocation UserLocation `json:"userLocation" validate:"required"`
	Orders       []Order      `json:"orders" validate:"required,min=1,dive"`
}

type EstimateResponse struct {
	TotalPrice                     int64  `json:"totalPrice"`
	EstimatedDeliveryTimeInMinutes int    `json:"estimatedDeliveryTimeInMinutes"`
	CalculatedEstimateId           string `json:"calculatedEstimateId"`
}

type CreateOrderRequest struct {
	CalculatedEstimateId string `json:"calculatedEstimateId" validate:"required"`
}

type CreateOrderResponse struct {
	OrderId string `json:"orderId"`
}

type MerchantPoint struct {
	MerchantID string
	Lat, Lng   float64
	IsStart    bool
	Order      Order
}

// OrderFilter holds filter params for get user orders
type OrderFilter struct {
	MerchantID       string
	Name             string
	MerchantCategory string
	Offset           int
	Limit            int
}

// Location represents geographical coordinates
type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

// Merchant represents merchant information in the order
type Merchant struct {
	MerchantID       string   `json:"merchantId"`
	Name             string   `json:"name"`
	MerchantCategory string   `json:"merchantCategory"`
	ImageURL         string   `json:"imageUrl"`
	Location         Location `json:"location"`
	CreatedAt        string   `json:"createdAt"`
}

// Item represents an item in the order
type Item struct {
	ItemID          string `json:"itemId"`
	Name            string `json:"name"`
	ProductCategory string `json:"productCategory"`
	Price           int64  `json:"price"`
	Quantity        int    `json:"quantity"`
	ImageURL        string `json:"imageUrl"`
	CreatedAt       string `json:"createdAt"`
}

// MerchantOrder represents orders from a specific merchant
type MerchantOrder struct {
	Merchant Merchant `json:"merchant"`
	Items    []Item   `json:"items"`
}

// OrderResponse represents a single order with all its merchants and items
type OrderResponse struct {
	OrderID string          `json:"orderId"`
	Orders  []MerchantOrder `json:"orders"`
}

// GetOrdersResponse is the final response structure
type GetOrdersResponse []OrderResponse

// Domain errors
var (
	ErrOrderNotFound        = errors.New("order not found")
	ErrInvalidMerchantID    = errors.New("invalid merchant id format")
	ErrInvalidCategory      = errors.New("invalid merchant category")
	ErrInvalidPagination    = errors.New("invalid pagination parameters")
	ErrUnauthorizedAccess   = errors.New("unauthorized to access this order")
	ErrDatabaseQuery        = errors.New("database query failed")
	ErrCacheOperation       = errors.New("cache operation failed")
	ErrDataTransformation   = errors.New("failed to transform data")
	ErrInvalidUserID        = errors.New("invalid user id")
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
		Error:   "validation_error",
		Message: message,
		Errors:  errors,
	}
}
