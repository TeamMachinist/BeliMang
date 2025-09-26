package items

import (
	"errors"
	"net/http"

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
