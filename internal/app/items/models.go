package items

import (
	"errors"

	"github.com/google/uuid"
)

type CreateItemRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=30"`
	ProductCategory string `json:"productCategory" validate:"required,oneof=Beverage Food Snack Condiments Additions"`
	Price           int64  `json:"price" validate:"required,min=1"`
	ImageUrl        string `json:"imageUrl" validate:"required,url"`
}

// items/types.go
type ListItemsRequest struct {
	ItemID          *uuid.UUID
	Name            *string
	ProductCategory *string
	CreatedAtOrder  *string
	Limit           int32
	Offset          int32
}

type ItemResponse struct {
	ItemID          string `json:"itemId"`
	Name            string `json:"name"`
	ProductCategory string `json:"productCategory"`
	Price           int64  `json:"price"`
	ImageUrl        string `json:"imageUrl"`
	CreatedAt       string `json:"createdAt"`
}

type ListItemsResponse struct {
	Data []ItemResponse `json:"data"`
	Meta struct {
		Limit  int32 `json:"limit"`
		Offset int32 `json:"offset"`
		Total  int64 `json:"total"`
	} `json:"meta"`
}

var ErrMerchantNotFound = errors.New("merchant not found")
