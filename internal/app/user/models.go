package user

import (
	"errors"
	"time"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

// User represents the user entity in the domain
type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password"`
	Email     string    `json:"email" db:"email"`
	Role      UserRole  `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RegisterRequest represents the request payload for registering a user
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=5,max=30"`
	Password string `json:"password" validate:"required,min=5,max=30"`
	Email    string `json:"email" validate:"required,email"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=5,max=30"`
	Password string `json:"password" validate:"required,min=5,max=30"`
}

// AuthResponse represents the response payload for auth operations
type AuthResponse struct {
	Token string `json:"token"`
}

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

// Domain errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Response constructors
func NewErrorResponse(err string, message string) *ErrorResponse {
	return &ErrorResponse{
		Error:   err,
		Message: message,
	}
}

func NewErrorResponseWithDetails(err string, message string, details map[string]string) *ErrorResponse {
	return &ErrorResponse{
		Error:   err,
		Message: message,
		Details: details,
	}
}

func NewValidationErrorResponse(message string, errors []ValidationError) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Error:   "validation_error",
		Message: message,
		Errors:  errors,
	}
}
