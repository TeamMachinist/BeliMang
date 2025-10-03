package purchase

type UserLocation struct {
	Lat  float64 `json:"lat" validate:"required,gte=-90,lte=90"`
	Long float64 `json:"long" validate:"required,gte=-180,lte=180"`
}

type OrderItem struct {
	ItemID   string `json:"itemId" validate:"required,uuid"`
	Quantity int    `json:"quantity" validate:"required,min=1"`
}

type Order struct {
	MerchantID      string      `json:"merchantId" validate:"required,uuid"`
	IsStartingPoint bool        `json:"isStartingPoint" validate:"required"`
	Items           []OrderItem `json:"items" validate:"required,min=1,dive"`
}

type EstimateRequest struct {
	UserLocation UserLocation `json:"userLocation" validate:"required"`
	Orders       []Order      `json:"orders" validate:"required,min=1,dive"`
}

type EstimateResponse struct {
	TotalPrice                     int64  `json:"totalPrice"`
	EstimatedDeliveryTimeInMinutes int    `json:"estimatedDeliveryTimeInMinutes"`
	CalculatedEstimateId           string `json:"calculatedEstimateId"`
}

type CreateOrderRequest struct {
	CalculatedEstimateId string `json:"calculatedEstimateId" validate:"required,uuid"`
}

type CreateOrderResponse struct {
	OrderId string `json:"orderId"`
}

type MerchantPoint struct {
	MerchantID string
	Lat, Lng   float64
	IsStart    bool
	Order      Order
}
// models :
type MerchantWithItemsResponse struct {
	Merchant MerchantInfo `json:"merchant"`
	Items    []ItemInfo   `json:"items"`
}

type MerchantInfo struct {
	MerchantID       string   `json:"merchantId"`
	Name             string   `json:"name"`
	MerchantCategory string   `json:"merchantCategory"`
	ImageUrl         string   `json:"imageUrl"`
	Location         Location `json:"location"`
	CreatedAt        string   `json:"createdAt"` // ISO 8601 with nanoseconds
}

type ItemInfo struct {
	ItemID          string `json:"itemId"`
	Name            string `json:"name"`
	ProductCategory string `json:"productCategory"`
	Price           int64  `json:"price"`
	ImageUrl        string `json:"imageUrl"`
	CreatedAt       string `json:"createdAt"` // ISO 8601 with nanoseconds
}

type Location struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

type GetMerchantsNearbyResponse struct {
	Data []MerchantWithItemsResponse `json:"data"`
	Meta PaginationMeta              `json:"meta"`
}

type PaginationMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}