package items

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ItemHandler struct {
	itemService *ItemService
}

func NewItemHandler(itemService *ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

func (h *ItemHandler) CreateItem(c *gin.Context) {
	merchantIDStr := c.Param("merchantId")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchantId format"})
		return
	}

	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, e := range validationErrors {
				field := e.Field()
				tag := e.Tag()

				switch tag {
				case "required":
					errorMessages = append(errorMessages, field+" is required")
				case "min":
					if field == "Price" {
						errorMessages = append(errorMessages, field+" must be at least "+e.Param())
					} else {
						errorMessages = append(errorMessages, field+" must be at least "+e.Param()+" characters")
					}
				case "max":
					errorMessages = append(errorMessages, field+" must not exceed "+e.Param()+" characters")
				case "oneof":
					errorMessages = append(errorMessages, field+" must be one of: Beverage, Food, Snack, Condiments, Additions")
				case "url":
					errorMessages = append(errorMessages, field+" must be a valid URL")
				default:
					errorMessages = append(errorMessages, "invalid "+field)
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errorMessages,
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON format"})
		return
	}

	itemID, err := h.itemService.CreateItem(c.Request.Context(), merchantID, req)
	if err != nil {
		switch {
		case errors.Is(err, errors.New("merchant not found")):
			c.JSON(http.StatusNotFound, gin.H{"error": "merchant not found"})
		case errors.Is(err, errors.New("validation failed")):
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"itemId": itemID.String(),
	})
}

func (h *ItemHandler) GetItems(c *gin.Context) {
	merchantIDStr := c.Param("merchantId")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchantId format"})
		return
	}

	req := ListItemsRequest{
		Limit:  5,
		Offset: 0,
	}

	if itemIdStr := c.Query("itemId"); itemIdStr != "" {
		itemID, err := uuid.Parse(itemIdStr)
		if err != nil {
			h.respondEmpty(c, 5, 0, 0)
			return
		}
		req.ItemID = &itemID
	}

	// --- limit & offset ---
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 {
			req.Limit = int32(l)
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil && o >= 0 {
			req.Offset = int32(o)
		}
	}

	// --- name (optional string) ---
	if name := c.Query("name"); name != "" {
		req.Name = &name
	}

	// --- productCategory (optional enum) ---
	if catStr := c.Query("productCategory"); catStr != "" {
		allowed := map[string]bool{
			"Beverage": true, "Food": true, "Snack": true,
			"Condiments": true, "Additions": true,
		}
		if allowed[catStr] {
			req.ProductCategory = &catStr
		}

	}

	// --- createdAt (optional: asc/desc) ---
	if order := c.Query("createdAt"); order != "" {
		lower := strings.ToLower(order)
		if lower == "asc" || lower == "desc" {
			req.CreatedAtOrder = &order
		}

	}

	data, total, err := h.itemService.ListItems(c.Request.Context(), merchantID, req)
	if err != nil {
		if errors.Is(err, ErrMerchantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "merchant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := ListItemsResponse{
		Data: data,
	}
	resp.Meta.Limit = req.Limit
	resp.Meta.Offset = req.Offset
	resp.Meta.Total = total

	c.JSON(http.StatusOK, resp)
}

func (h *ItemHandler) respondEmpty(c *gin.Context, limit, offset int32, total int64) {
	c.JSON(http.StatusOK, ListItemsResponse{
		Data: []ItemResponse{},
		Meta: struct {
			Limit  int32 `json:"limit"`
			Offset int32 `json:"offset"`
			Total  int64 `json:"total"`
		}{Limit: limit, Offset: offset, Total: 0},
	})
}
