package purchase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PurchaseHandler struct {
	purchaseService *PurchaseService
}

func NewPurchaseHandler(pS *PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{purchaseService: pS}
}

func (h *PurchaseHandler) Estimate(c *gin.Context) {

	var req EstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	resp, err := h.purchaseService.ValidateAndEstimate(c, req)
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
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Parse the estimate ID from the request
	estimateID, err := uuid.Parse(req.CalculatedEstimateId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid estimate ID"})
		return
	}

	resp, err := h.purchaseService.CreateOrderByEstimateId(c, estimateID)
	if err != nil {
		switch err.Error() {
		case "estimate not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "estimate not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
