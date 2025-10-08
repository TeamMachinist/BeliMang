package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UserHandler handles user HTTP requests
type UserHandler struct {
	service  *UserService
	validate *validator.Validate
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(service *UserService, validate *validator.Validate) *UserHandler {
	return &UserHandler{
		service:  service,
		validate: validate,
	}
}

// RegisterUser creates a new user account
func (h *UserHandler) RegisterUser(c *gin.Context) {
	h.register(c, UserRoleUser)
}

// RegisterAdmin creates a new admin account
func (h *UserHandler) RegisterAdmin(c *gin.Context) {
	h.register(c, UserRoleAdmin)
}

// LoginUser authenticates a user
func (h *UserHandler) LoginUser(c *gin.Context) {
	h.login(c, UserRoleUser)
}

// LoginAdmin authenticates an admin
func (h *UserHandler) LoginAdmin(c *gin.Context) {
	h.login(c, UserRoleAdmin)
}

// register handles registration for both users and admins
func (h *UserHandler) register(c *gin.Context, role UserRole) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("validation_error", err.Error()))
		return
	}

	// Validate request using validator library
	if err := h.validate.Struct(req); err != nil {
		var validationErrors []ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Message: getValidationMessage(err),
				Value:   getFieldValue(err),
			})
		}
		response := NewValidationErrorResponse("Validation failed", validationErrors)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	token, err := h.service.Register(&req, role)
	if err != nil {
		switch err {
		case ErrUsernameExists:
			c.JSON(http.StatusConflict, NewErrorResponse("username_conflict", "Username already exists"))
			return
		case ErrEmailExists:
			c.JSON(http.StatusConflict, NewErrorResponse("email_conflict", "Email already exists"))
			return
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse("internal_error", err.Error()))
			return
		}
	}

	c.JSON(http.StatusCreated, &AuthResponse{Token: token})
}

// login handles login for both users and admins
func (h *UserHandler) login(c *gin.Context, role UserRole) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("validation_error", err.Error()))
		return
	}

	// Validate request using validator library
	if err := h.validate.Struct(req); err != nil {
		var validationErrors []ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Message: getValidationMessage(err),
				Value:   getFieldValue(err),
			})
		}
		response := NewValidationErrorResponse("Validation failed", validationErrors)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	token, err := h.service.Login(&req, role)
	if err != nil {
		if err == ErrInvalidCredentials {
			c.JSON(http.StatusBadRequest, NewErrorResponse("invalid_credentials", "Invalid username or password"))
			return
		}
		c.JSON(http.StatusInternalServerError, NewErrorResponse("internal_error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, &AuthResponse{Token: token})
}

// getValidationMessage returns a human-readable validation message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short (minimum " + err.Param() + " characters)"
	case "max":
		return "Value is too long (maximum " + err.Param() + " characters)"
	default:
		return "Invalid value"
	}
}

// getFieldValue safely extracts field value for validation errors
func getFieldValue(err validator.FieldError) string {
	if err.Value() == nil {
		return ""
	}
	if str, ok := err.Value().(string); ok {
		return str
	}
	return ""
}
