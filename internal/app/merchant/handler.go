package merchant

import (
	"net/http"
	"strconv"
	"strings"

	logger "belimang/internal/pkg/logging"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type MerchantHandler struct {
	service  *MerchantService
	validate *validator.Validate
}

func NewMerchantHandler(service *MerchantService, validate *validator.Validate) *MerchantHandler {
	validate.RegisterValidation("merchantCategory", MerchantCategoryValidator)
	validate.RegisterValidation("urlSuffix", imageURLValidator)

	return &MerchantHandler{
		service:  service,
		validate: validate,
	}
}

func (h *MerchantHandler) CreateMerchantHandler(c *gin.Context) {
	adminID, err := getUserID(c)
	if err != nil {
		logger.ErrorCtx(c, "Unauthorized account", "error", err.Error())
		c.JSON(http.StatusUnauthorized, NewErrorResponse("unathorized error", err.Error()))
		return
	}

	var req PostMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("validation error", err.Error()))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		var validationErrors []ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Message: getValidationMessage(err),
				Value:   getFieldValue(err),
			})
		}
		resp := NewValidationErrorResponse("Validation failed", validationErrors)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchantID, err := h.service.CreateMerchantService(c, adminID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("internal error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, merchantID)
}

func (h *MerchantHandler) SearchMerchantsHandler(c *gin.Context) {
	filter := MerchantFilter{
		MerchantID:       c.Query("merchantId"),
		Name:             c.Query("name"),
		MerchantCategory: c.Query("merchantCategory"),
		CreatedAtSort:    c.DefaultQuery("createdAt", "desc"),
	}
	// Parse limit & offset
	if l := c.DefaultQuery("limit", "5"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			filter.Limit = v
		} else {
			filter.Limit = 5
		}
	} else {
		filter.Limit = 5
	}
	if o := c.DefaultQuery("offset", "0"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			filter.Offset = v
		} else {
			filter.Offset = 0
		}
	} else {
		filter.Offset = 0
	}

	// Validate merchantCategory
	if filter.MerchantCategory != "" {
		if _, ok := validMerchantCategories[filter.MerchantCategory]; !ok {
			c.JSON(http.StatusOK, GetMerchantsResponse{Data: []Merchant{}, Meta: Meta{Limit: filter.Limit, Offset: filter.Offset, Total: 0}})
			return
		}
	}
	// Validate createdAtSort
	if filter.CreatedAtSort != "asc" && filter.CreatedAtSort != "desc" {
		filter.CreatedAtSort = "desc"
	}

	resp, err := h.service.SearchMerchantsService(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("internal error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp)
}

var validMerchantCategories = map[string]struct{}{
	"SmallRestaurant":       {},
	"MediumRestaurant":      {},
	"LargeRestaurant":       {},
	"MerchandiseRestaurant": {},
	"BoothKiosk":            {},
	"ConvenienceStore":      {},
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	rawUserID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, ErrUserNotFound
	}
	userID, _ := uuid.Parse(rawUserID.(string))

	return userID, nil
}

func MerchantCategoryValidator(fl validator.FieldLevel) bool {
	_, exists := validMerchantCategories[fl.Field().String()]
	return exists
}

func imageURLValidator(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	return strings.HasSuffix(url, ".jpg") || strings.HasSuffix(url, ".jpeg")
}

// getValidationMessage returns a human-readable validation message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "url":
		return "Must be a valid URL"
	case "min":
		return "Value is too short (minimum " + err.Param() + " characters)"
	case "max":
		return "Value is too long (maximum " + err.Param() + " characters)"
	case "latitude":
		return "Latitude must be between -90 and 90"
	case "longitude":
		return "Longitude must be between -180 and 180"
	default:
		return "Invalid value"
	}
}

func getFieldValue(err validator.FieldError) string {
	if err.Value() == nil {
		return ""
	}
	if str, ok := err.Value().(string); ok {
		return str
	}
	return ""
}
