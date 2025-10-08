package purchase

import (
	logger "belimang/internal/pkg/logging"
	"net/http"
	"strconv"
	"strings"

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

func (h *PurchaseHandler) GetMerchantsNearbyHandler(c *gin.Context) {
	// Parse coordinates
	coords := c.Param("coords")
	parts := strings.Split(coords, ",")

	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coordinates format. Use lat,lng"})
		return
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
		return
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
		return
	}

	// // Validate lat/lng ranges
	// if lat < -90 || lat > 90 {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "latitude must be between -90 and 90"})
	// 	return
	// }

	// if lng < -180 || lng > 180 {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "longitude must be between -180 and 180"})
	// 	return
	// }

	// Parse query parameters
	merchantID := c.DefaultQuery("merchantId", "")
	name := c.DefaultQuery("name", "")
	merchantCategory := c.DefaultQuery("merchantCategory", "")

	// Parse limit & offset
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a valid positive number"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be a valid positive number"})
		return
	}

	// Validate merchantCategory enum (if provided)
	validCategories := map[string]bool{
		"SmallRestaurant":       true,
		"MediumRestaurant":      true,
		"LargeRestaurant":       true,
		"MerchandiseRestaurant": true,
		"BoothKiosk":            true,
		"ConvenienceStore":      true,
	}

	if merchantCategory != "" && !validCategories[merchantCategory] {
		// Return 200 with empty array for invalid category
		c.JSON(http.StatusOK, GetMerchantsNearbyResponse{
			Data: []MerchantWithItemsResponse{},
			Meta: PaginationMeta{
				Limit:  limit,
				Offset: offset,
				Total:  0,
			},
		})
		return
	}

	// Validate merchantId format (if provided)
	if merchantID != "" {
		if _, err := uuid.Parse(merchantID); err != nil {
			// Return 200 with empty array for invalid UUID
			c.JSON(http.StatusOK, GetMerchantsNearbyResponse{
				Data: []MerchantWithItemsResponse{},
				Meta: PaginationMeta{
					Limit:  limit,
					Offset: offset,
					Total:  0,
				},
			})
			return
		}
	}

	// Call service
	ctx := c.Request.Context()
	response, err := h.purchaseService.GetMerchantsNearby(ctx, &GetMerchantsNearbyParams{
		Lat:              lat,
		Lng:              lng,
		MerchantID:       merchantID,
		Name:             name,
		MerchantCategory: merchantCategory,
		Limit:            limit,
		Offset:           offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Always return 200, even if empty
	c.JSON(http.StatusOK, response)
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
