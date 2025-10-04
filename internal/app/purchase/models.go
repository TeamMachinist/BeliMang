package purchase

type UserLocation struct {
	Lat  float64 `json:"lat" validate:"required,gte=-90,lte=90"`
	Long float64 `json:"long" validate:"required,gte=-180,lte=180"`
}

type OrderItem struct {
	ItemID   string `json:"itemId" validate:"required"`
	Quantity int    `json:"quantity" validate:"required,min=1"`
}

type Order struct {
	MerchantID      string      `json:"merchantId" validate:"required"`
	IsStartingPoint bool        `json:"isStartingPoint"`
	Items           []OrderItem `json:"items" validate:"dive"`
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
	CalculatedEstimateId string `json:"calculatedEstimateId" validate:"required"`
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
