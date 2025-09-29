package purchase

type UserLocation struct {
	Lat  float64 `json:"lat" validate:"required,gte=-90,lte=90"`
	Long float64 `json:"long" validate:"required,gte=-180,lte=180"`
}

type OrderItem struct {
	ItemID   string `json:"itemId" validate:"required,uuid"`
	Quantity int64  `json:"quantity" validate:"required,min=1"`
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
