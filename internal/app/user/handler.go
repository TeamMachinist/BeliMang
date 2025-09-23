package user

import (
	"net/http"
	"strconv"

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

// Create creates a new user
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
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
				Value:   err.Value().(string),
			})
		}
		response := NewValidationErrorResponse("Validation failed", validationErrors)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := h.service.Create(&req)
	if err != nil {
		if err == ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, NewErrorResponse("user_exists", "User with this email already exists"))
			return
		}
		c.JSON(http.StatusInternalServerError, NewErrorResponse("internal_error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetAll retrieves all users
func (h *UserHandler) GetAll(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Ensure limit is reasonable
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	users, err := h.service.GetAll(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("internal_error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetByID retrieves a user by ID
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("invalid_request", "User ID is required"))
		return
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		if err == ErrUserNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("user_not_found", "User not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, NewErrorResponse("internal_error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update updates an existing user
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("invalid_request", "User ID is required"))
		return
	}

	var req UpdateUserRequest
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
				Value:   err.Value().(string),
			})
		}
		response := NewValidationErrorResponse("Validation failed", validationErrors)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := h.service.Update(id, &req)
	if err != nil {
		if err == ErrUserNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("user_not_found", "User not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, NewErrorResponse("internal_error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete removes a user by ID
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("invalid_request", "User ID is required"))
		return
	}

	err := h.service.Delete(id)
	if err != nil {
		if err == ErrUserNotFound {
			c.JSON(http.StatusNotFound, NewErrorResponse("user_not_found", "User not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, NewErrorResponse("internal_error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// getValidationMessage returns a human-readable validation message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "omitempty":
		return "Invalid value"
	default:
		return "Invalid value"
	}
}
