package purchase

import (
	"context"
	// "encoding/json"
	"errors"
	"fmt"
	"time"

	// "belimang/internal/infrastructure/cache"
	"belimang/internal/infrastructure/database"
	logger "belimang/internal/pkg/logging"
	"belimang/internal/pkg/utils"

	"github.com/google/uuid"
	"github.com/uber/h3-go/v4"
)

var ErrNeedExactValidation = errors.New("ambiguous distance: need exact validation")

type PurchaseService struct {
	queries *database.Queries
	db      *database.DB
}

func NewPurchaseService(q *database.Queries, db *database.DB) *PurchaseService {
	return &PurchaseService{
		queries: q,
		db:      db,
	}
}

// validateWithH3PreFilter returns:
// - nil → all merchants ≤3000m
// - error → any merchant >3000m
func (s *PurchaseService) validateWithH3PreFilter(
	userH3 h3.Cell,
	merchantPoints []merchantPoint,
) error {
	for _, mp := range merchantPoints {
		dist := utils.H3GridDistanceMeters(userH3, mp.H3Cell)
		if dist < 0 {
			return errors.New("coordinates too far")
		}
		if dist > 3000 {
			return errors.New("coordinates too far")
		}
	}
	return nil
}

type merchantPoint struct {
	MerchantID string
	Lat, Lng   float64
	H3Cell     h3.Cell
	IsStart    bool
	Order      Order
}

func (s *PurchaseService) ValidateAndEstimate(ctx context.Context, userID uuid.UUID, req EstimateRequest) (EstimateResponse, error) {
	if len(req.Orders) == 0 {
		return EstimateResponse{}, errors.New("orders cannot be empty")
	}

	var merchantIDs []uuid.UUID
	merchantIdMap := make(map[string]uuid.UUID)
	for _, o := range req.Orders {
		parsedMerchantID, err := uuid.Parse(o.MerchantID)
		if err != nil {
			return EstimateResponse{}, errors.New("merchant not found")
		}
		merchantIDs = append(merchantIDs, parsedMerchantID)
		merchantIdMap[o.MerchantID] = parsedMerchantID
	}

	merchants, err := s.queries.GetMerchantsLatLong(ctx, merchantIDs)
	if err != nil {
		return EstimateResponse{}, errors.New("failed to fetch merchants")
	}

	if len(merchants) != len(merchantIDs) {
		return EstimateResponse{}, errors.New("merchant not found")
	}

	merchantMap := make(map[uuid.UUID]database.GetMerchantsLatLongRow)
	for _, m := range merchants {
		merchantMap[m.ID] = m
	}

	startCount := 0
	for _, o := range req.Orders {
		if o.IsStartingPoint {
			startCount++
		}
	}
	if startCount != 1 {
		return EstimateResponse{}, errors.New("exactly one order must have isStartingPoint=true")
	}

	var points []merchantPoint
	totalPrice := int(0)

	userH3, err := utils.LatLonToH3(req.UserLocation.Lat, req.UserLocation.Long)
	if err != nil {
		return EstimateResponse{}, errors.New("invalid user location")
	}

	var itemIDs []uuid.UUID
	var itemMerchantIDs []uuid.UUID
	itemQuantities := make(map[string]int)

	for _, o := range req.Orders {
		parsedMerchantID := merchantIdMap[o.MerchantID]
		for _, item := range o.Items {
			parsedItemID, err := uuid.Parse(item.ItemID)
			if err != nil {
				return EstimateResponse{}, errors.New("item not found")
			}
			itemIDs = append(itemIDs, parsedItemID)
			itemMerchantIDs = append(itemMerchantIDs, parsedMerchantID)

			key := parsedItemID.String() + "-" + parsedMerchantID.String()
			itemQuantities[key] = item.Quantity
		}
	}

	itemPrices, err := s.queries.GetItemPricesByIDsAndMerchants(ctx, database.GetItemPricesByIDsAndMerchantsParams{
		ItemID:     itemIDs,
		MerchantID: itemMerchantIDs,
	})
	if err != nil {
		return EstimateResponse{}, errors.New("failed to fetch item prices")
	}

	if len(itemPrices) != len(itemIDs) {
		return EstimateResponse{}, errors.New("item not found")
	}

	foundItems := make(map[string]bool)
	for _, itemPrice := range itemPrices {
		key := itemPrice.ID.String() + "-" + itemPrice.MerchantID.String()
		foundItems[key] = true
		if quantity, exists := itemQuantities[key]; exists {
			totalPrice += int(itemPrice.Price) * quantity
		}
	}

	for key := range itemQuantities {
		if !foundItems[key] {
			return EstimateResponse{}, errors.New("item not found")
		}
	}

	for _, o := range req.Orders {
		parsedMerchantID := merchantIdMap[o.MerchantID]
		merchant, exists := merchantMap[parsedMerchantID]
		if !exists {
			return EstimateResponse{}, errors.New("merchant not found")
		}

		h3Cell, err := utils.LatLonToH3(merchant.Lat, merchant.Lng)
		if err != nil {
			return EstimateResponse{}, errors.New("invalid merchant location")
		}

		points = append(points, merchantPoint{
			MerchantID: o.MerchantID,
			Lat:        merchant.Lat,
			Lng:        merchant.Lng,
			H3Cell:     h3Cell,
			IsStart:    o.IsStartingPoint,
			Order:      o,
		})
	}

	err = s.validateWithH3PreFilter(userH3, points)
	if err != nil {
		return EstimateResponse{}, errors.New("coordinates too far")
	}

	// Always use H3 for TSP since we've validated with H3
	const useHaversineForTSP = false

	var start *merchantPoint
	rest := make([]merchantPoint, 0)
	for i := range points {
		if points[i].IsStart {
			start = &points[i]
		} else {
			rest = append(rest, points[i])
		}
	}
	if start == nil {
		return EstimateResponse{}, errors.New("starting point not found")
	}

	route := []merchantPoint{*start}
	current := *start

	for len(rest) > 0 {
		bestIdx := -1
		if useHaversineForTSP {
			bestDist := 1e15
			for i, p := range rest {
				d := utils.HaversineDistance(current.Lat, current.Lng, p.Lat, p.Lng)
				if d < bestDist {
					bestDist = d
					bestIdx = i
				}
			}
		} else {
			bestGridDist := int(^uint(0) >> 1)
			for i, p := range rest {
				if d, err := h3.GridDistance(current.H3Cell, p.H3Cell); err == nil && d < bestGridDist {
					bestGridDist = d
					bestIdx = i
				}
			}
		}

		if bestIdx == -1 {
			break
		}
		route = append(route, rest[bestIdx])
		current = rest[bestIdx]
		rest = append(rest[:bestIdx], rest[bestIdx+1:]...)
	}

	// === Hitung total jarak ===
	totalDist := 0.0
	if useHaversineForTSP {
		for i := 0; i < len(route)-1; i++ {
			totalDist += utils.HaversineDistance(route[i].Lat, route[i].Lng, route[i+1].Lat, route[i+1].Lng)
		}
		last := route[len(route)-1]
		totalDist += utils.HaversineDistance(last.Lat, last.Lng, req.UserLocation.Lat, req.UserLocation.Long)
	} else {
		totalGridDist := 0
		for i := 0; i < len(route)-1; i++ {
			if d, err := h3.GridDistance(route[i].H3Cell, route[i+1].H3Cell); err == nil {
				totalGridDist += d
			}
		}
		if d, err := h3.GridDistance(route[len(route)-1].H3Cell, userH3); err == nil {
			totalGridDist += d
		}
		totalDist = float64(totalGridDist) * utils.EDGE_LENGTH_METERS
	}

	timeMinutes := utils.EstimateTimeMinutes(totalDist)

	repository := NewPurchaseRepository(s.db)
	estimate, err := repository.CreateEstimateWithOrders(ctx,
		userID,
		req.UserLocation.Lat,
		req.UserLocation.Long,
		float64(totalPrice),
		timeMinutes,
		req.Orders)
	if err != nil {
		return EstimateResponse{}, fmt.Errorf("failed to save estimate: %w", err)
	}

	return EstimateResponse{
		TotalPrice:                     estimate.TotalPrice,
		EstimatedDeliveryTimeInMinutes: int(estimate.EstimatedDeliveryTimeInMinutes),
		CalculatedEstimateId:           estimate.ID.String(),
	}, nil
}

func (s *PurchaseService) CreateOrderByEstimateId(ctx context.Context, userID uuid.UUID, estimateID uuid.UUID) (CreateOrderResponse, error) {
	repository := NewPurchaseRepository(s.db)

	_, err := repository.GetEstimateById(ctx, estimateID)
	if err != nil {
		return CreateOrderResponse{}, errors.New("estimate not found")
	}

	order, err := repository.CreateOrderFromEstimate(ctx, userID, estimateID)
	if err != nil {
		return CreateOrderResponse{}, fmt.Errorf("failed to create order from estimate: %w", err)
	}

	return CreateOrderResponse{
		OrderId: order.ID.String(),
	}, nil
}

func (s *PurchaseService) GetOrdersService(ctx context.Context, userID uuid.UUID, filter OrderFilter) (GetOrdersResponse, error) {
	// Validate user ID
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	var merchantID uuid.UUID
	if filter.MerchantID != "" {
		id, err := uuid.Parse(filter.MerchantID)
		if err == nil {
			merchantID = id
		}
	}

	// Sets default values for optional fields
	if filter.Limit == 0 {
		filter.Limit = 5
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Generate cache key
	// cacheKey := s.generateCacheKey(userID, filter)

	// Try to get from cache
	// cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	// if err == nil {
	// 	var result GetOrdersResponse
	// 	if unmarshalErr := json.Unmarshal([]byte(cachedData), &result); unmarshalErr == nil {
	// 		return result, nil
	// 	}
	// 	// Log cache unmarshal error but continue with database query
	// 	log.Printf("Failed to unmarshal cached data: %v", unmarshalErr)
	// } else if err != redis.Nil {
	// 	// Log cache read error but continue with database query
	// 	log.Printf("Redis get error: %v", err)
	// }

	// Prepare query parameters
	// var namePattern *string
	// if filter.Name != nil && *filter.Name != "" {
	// 	pattern := fmt.Sprintf("%%%s%%", *filter.Name)
	// 	namePattern = &pattern
	// }

	// Execute query
	repository := NewPurchaseRepository(s.db)
	rows, err := repository.GetOrders(
		ctx, userID, merchantID, filter.Name, filter.MerchantCategory, int32(filter.Limit), int32(filter.Offset))
	if err != nil {
		return GetOrdersResponse{}, ErrDatabaseQuery
	}

	// Handle empty result
	// if len(rows) == 0 {
	// 	// Cache empty result for a shorter duration (1 minute)
	// 	emptyResult := GetOrdersResponse{}
	// 	if jsonData, err := json.Marshal(emptyResult); err == nil {
	// 		if setErr := s.redis.Set(ctx, cacheKey, jsonData, 1*time.Minute).Err(); setErr != nil {
	// 			log.Printf("Failed to cache empty result: %v", setErr)
	// 		}
	// 	}
	// 	return emptyResult, nil
	// }

	// Transform the flat result into nested structure
	result, err := s.transformToOrderResponse(rows)
	if err != nil {
		logger.ErrorCtx(ctx, "Data transformation error", err)
		return nil, ErrDataTransformation
	}

	// Cache the result for 5 minutes
	// if len(result) > 0 {
	// 	jsonData, err := json.Marshal(result)
	// 	if err != nil {
	// 		log.Printf("Failed to marshal result for caching: %v", err)
	// 		// Don't return error, just skip caching
	// 	} else {
	// 		if setErr := s.redis.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err(); setErr != nil {
	// 			log.Printf("Failed to cache result: %v", setErr)
	// 			// Don't return error, caching is optional
	// 		}
	// 	}
	// }

	return result, nil
}

func (s *PurchaseService) transformToOrderResponse(rows []database.GetOrdersWithDetailsRow) (GetOrdersResponse, error) {
	if len(rows) == 0 {
		return GetOrdersResponse{}, nil
	}

	// Track unique order IDs and merchant IDs per order
	type orderKey struct {
		orderID    uuid.UUID
		merchantID uuid.UUID
	}

	orderMap := make(map[uuid.UUID]int)      // orderID -> index in result
	merchantMap := make(map[orderKey]int)    // composite key -> index in order.Orders
	result := make(GetOrdersResponse, 0, 16) // Pre-allocate reasonable capacity

	for i := range rows {
		row := &rows[i]

		// Get or create order
		orderIdx, orderExists := orderMap[row.OrderID]
		if !orderExists {
			orderIdx = len(result)
			orderMap[row.OrderID] = orderIdx
			result = append(result, OrderResponse{
				OrderID: row.OrderID.String(),
				Orders:  make([]MerchantOrder, 0, 4), // Pre-allocate for typical order
			})
		}

		// Get or create merchant order
		key := orderKey{orderID: row.OrderID, merchantID: row.MerchantID}
		merchantIdx, merchantExists := merchantMap[key]
		if !merchantExists {
			merchantIdx = len(result[orderIdx].Orders)
			merchantMap[key] = merchantIdx
			result[orderIdx].Orders = append(result[orderIdx].Orders, MerchantOrder{
				Merchant: Merchant{
					MerchantID:       row.MerchantID.String(),
					Name:             row.MerchantName,
					MerchantCategory: row.MerchantCategory,
					ImageURL:         row.MerchantImageUrl,
					Location: Location{
						Latitude:  row.MerchantLat,
						Longitude: row.MerchantLng,
					},
					CreatedAt: formatTimestamp(row.MerchantCreatedAt),
				},
				Items: make([]Item, 0, 8), // Pre-allocate for typical item count
			})
		}

		// Append item directly
		result[orderIdx].Orders[merchantIdx].Items = append(
			result[orderIdx].Orders[merchantIdx].Items,
			Item{
				ItemID:          row.ItemID.String(),
				Name:            row.ItemName,
				ProductCategory: row.ProductCategory,
				Price:           row.ItemPrice,
				Quantity:        row.Quantity,
				ImageURL:        row.ItemImageUrl,
				CreatedAt:       formatTimestamp(row.ItemCreatedAt),
			},
		)
	}

	return result, nil
}

// func (s *PurchaseService) generateCacheKey(userID uuid.UUID, filter FilterOrderRequest) string {
// 	key := fmt.Sprintf("orders:user:%s:limit:%d:offset:%d", userID.String(), filter.Limit, filter.Offset)

// 	if filter.MerchantID != nil {
// 		key += fmt.Sprintf(":merchant:%s", filter.MerchantID.String())
// 	}

// 	if filter.Name != nil && *filter.Name != "" {
// 		key += fmt.Sprintf(":name:%s", *filter.Name)
// 	}

// 	if filter.MerchantCategory != nil && *filter.MerchantCategory != "" {
// 		key += fmt.Sprintf(":category:%s", *filter.MerchantCategory)
// 	}

// 	return key
// }

// InvalidateOrderCache invalidates the cache for a specific user
// func (s *OrderService) InvalidateOrderCache(ctx context.Context, userID uuid.UUID) error {
// 	pattern := fmt.Sprintf("orders:user:%s:*", userID.String())

// 	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()
// 	for iter.Next(ctx) {
// 		if err := s.redis.Del(ctx, iter.Val()).Err(); err != nil {
// 			log.Printf("Failed to delete cache key %s: %v", iter.Val(), err)
// 		}
// 	}

// 	if err := iter.Err(); err != nil {
// 		return fmt.Errorf("%w: %v", ErrCacheOperation, err)
// 	}

// 	return nil
// }

func formatTimestamp(t time.Time) string {
	// Format in ISO 8601 with nanoseconds
	return t.Format(time.RFC3339Nano)
}
