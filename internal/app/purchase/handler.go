package purchase

import (
	"net/http"
	"strconv"

	logger "belimang/internal/pkg/logging"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type PurchaseHandler struct {
	purchaseService *PurchaseService
	validate        *validator.Validate
}

func NewPurchaseHandler(pS *PurchaseService, v *validator.Validate) *PurchaseHandler {
	return &PurchaseHandler{purchaseService: pS, validate: v}
}

func (h *PurchaseHandler) Estimate(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
		return
	}

	var req EstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := h.purchaseService.ValidateAndEstimate(c, userUUID, req)
	if err != nil {
		switch err.Error() {
		case "merchant not found", "item not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case "coordinates too far":
			c.JSON(http.StatusBadRequest, gin.H{"error": "coordinates too far"})
		case "exactly one order must have isStartingPoint=true",
			"orders cannot be empty",
			"starting point not found":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PurchaseHandler) CreateOrder(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	estimateID, err := uuid.Parse(req.CalculatedEstimateId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid estimate ID"})
		return
	}

	resp, err := h.purchaseService.CreateOrderByEstimateId(c, userUUID, estimateID)
	if err != nil {
		switch err.Error() {
		case "estimate not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "estimate not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *PurchaseHandler) GetOrdersHandler(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		logger.ErrorCtx(c, "Unauthorized account", "error", err.Error())
		c.JSON(http.StatusUnauthorized, NewErrorResponse("unathorized error", err.Error()))
		return
	}

	filter := OrderFilter{
		MerchantID:       c.Query("merchantId"),
		Name:             c.Query("name"),
		MerchantCategory: c.Query("merchantCategory"),
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
			c.JSON(http.StatusOK, GetOrdersResponse{})
			return
		}
	}

	resp, err := h.purchaseService.GetOrdersService(c, userID, filter)
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
		return uuid.Nil, ErrInvalidUserID
	}
	userID, _ := uuid.Parse(rawUserID.(string))

	return userID, nil
}
