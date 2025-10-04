package purchase

import (
	"net/http"

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
