package purchase

import (
	"context"
	"net/http"
	"strconv"

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

// GetMerchantsNearbyHandler handles GET /merchants/nearby
func (h *PurchaseHandler) GetMerchantsNearbyHandler(c *gin.Context) {
	// Parse query parameters: lat and lng
	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing required query parameters: 'lat' and 'lng'",
		})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid 'lat' parameter: must be a valid float",
		})
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid 'lng' parameter: must be a valid float",
		})
		return
	}

	// Validate coordinate ranges (optional but recommended)
	if lat < -90 || lat > 90 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "'lat' must be between -90 and 90",
		})
		return
	}
	if lng < -180 || lng > 180 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "'lng' must be between -180 and 180",
		})
		return
	}

	// Call service
	ctx := context.Background() // or use c.Request.Context() if you have timeouts/tracing
	response, err := h.purchaseService.GetMerchantsNearby(ctx, lat, lng)
	if err != nil {
		// Log the error internally
		// logger.Error("Failed to get nearby merchants", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch nearby merchants",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, response)
}