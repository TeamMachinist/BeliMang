package purchase

import (
	"belimang/internal/infrastructure/database"
	"belimang/internal/pkg/utils"
	"context"
	"errors"
	"fmt"

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

func (s *PurchaseService) ValidateAndEstimate(ctx context.Context, req EstimateRequest) (EstimateResponse, error) {
	if len(req.Orders) == 0 {
		return EstimateResponse{}, errors.New("orders cannot be empty")
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
	totalPrice := int64(0)

	userH3, err := utils.LatLonToH3(req.UserLocation.Lat, req.UserLocation.Long)
	if err != nil {
		return EstimateResponse{}, errors.New("invalid user location")
	}

	var merchantIDs []uuid.UUID
	merchantIdMap := make(map[string]uuid.UUID)
	for _, o := range req.Orders {
		parsedMerchantID, _ := uuid.Parse(o.MerchantID)
		merchantIDs = append(merchantIDs, parsedMerchantID)
		merchantIdMap[o.MerchantID] = parsedMerchantID
	}

	merchants, err := s.queries.GetMerchantsLatLong(ctx, merchantIDs)
	if err != nil {
		return EstimateResponse{}, errors.New("failed to fetch merchants")
	}

	merchantMap := make(map[uuid.UUID]database.GetMerchantsLatLongRow)
	for _, m := range merchants {
		merchantMap[m.ID] = m
	}

	var itemIDs []uuid.UUID
	var itemMerchantIDs []uuid.UUID
	itemQuantities := make(map[string]int64)

	for _, o := range req.Orders {
		parsedMerchantID := merchantIdMap[o.MerchantID]
		for _, item := range o.Items {
			parsedItemID, _ := uuid.Parse(item.ItemID)
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

	for _, itemPrice := range itemPrices {
		key := itemPrice.ID.String() + "-" + itemPrice.MerchantID.String()
		if quantity, exists := itemQuantities[key]; exists {
			totalPrice += itemPrice.Price * quantity
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
		// Gunakan H3 untuk estimasi jarak total
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
		req.UserLocation.Lat,
		req.UserLocation.Long,
		float64(totalPrice),
		int32(timeMinutes),
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

func (s *PurchaseService) CreateOrderByEstimateId(ctx context.Context, estimateID uuid.UUID) (CreateOrderResponse, error) {
	repository := NewPurchaseRepository(s.db)

	_, err := repository.GetEstimateById(ctx, estimateID)
	if err != nil {
		return CreateOrderResponse{}, errors.New("estimate not found")
	}

	order, err := repository.CreateOrderFromEstimate(ctx, estimateID)
	if err != nil {
		return CreateOrderResponse{}, fmt.Errorf("failed to create order from estimate: %w", err)
	}

	return CreateOrderResponse{
		OrderId: order.ID.String(),
	}, nil
}

// func (s *PurchaseService) getNearestMerchant(ctx context.Context, lat float64, lng float64) (merchantList, error) {

// 	// 1. Generate H3 cell (resolusi 10 contoh)
// 	resolution := 10
// 	centerCell := h3.FromGeo(h3.GeoCoord{Lat: lat, Lng: lng}, resolution)

// 	// 2. Cari merchant di H3 cell dan tetangganya
// 	var candidates []Merchant
// 	maxRings := 5 // batas pencarian
// 	found := false

// 	for ring := 0; ring <= maxRings && !found; ring++ {
// 		neighbors := h3.KRing(centerCell, ring)
// 		// Konversi []h3.Cell ke []int64
// 		cellIDs := make([]int64, len(neighbors))
// 		for i, cell := range neighbors {
// 			cellIDs[i] = int64(cell)
// 		}

// 		// Query ke DB
// 		query := `
// 			SELECT id, name, lat, lng
// 			FROM merchants
// 			WHERE h3_cell = ANY($1)
// 		`
		
// 		rows, err := s.db.Pool.Query(context.Background(), query, cellIDs)
// 		if err != nil {
// 			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "db query failed"})
// 			return
// 		}
// 		defer rows.Close()

// 		for rows.Next() {
// 			var m Merchant
// 			err := rows.Scan(&m.ID, &m.Name, &m.Lat, &m.Lng)
// 			if err != nil {
// 				continue
// 			}
// 			candidates = append(candidates, m)
// 		}

// 		if len(candidates) > 0 {
// 			found = true
// 		}
// 	}

// 	if len(candidates) == 0 {
// 		// c.JSON(http.StatusNotFound, gin.H{"message": "no merchants found"})
// 		// return
// 	}

// 	// 3. Hitung jarak & urutkan
// 	type MerchantWithDistance struct {
// 		Merchant
// 		Distance float64
// 	}
// 	var results []MerchantWithDistance
// 	for _, m := range candidates {
// 		results = append(results, MerchantWithDistance{m, dist})
// 	}

// 	// Sort by distance
// 	sort.Slice(results, func(i, j int) bool {
// 		return results[i].Distance < results[j].Distance
// 	})

// 	// Ambil 1 terdekat (atau bisa ambil beberapa)
// 	nearest := results[0]

// 	c.JSON(http.StatusOK, gin.H{
// 		"merchant": nearest.Merchant,
// 		"distance_meters": nearest.Distance,
// 	})
// }

func (s *PurchaseService) GetMerchantsNearby(ctx context.Context, lat, lng float64) (GetMerchantsNearbyResponse, error) {
	rows, err := s.queries.GetAllMerchantsWithItemsSortedByH3Distance(ctx, database.GetAllMerchantsWithItemsSortedByH3DistanceParams{Point: lat, Point_2: lng})
	if err != nil {
		return GetMerchantsNearbyResponse{}, fmt.Errorf("failed to fetch merchants with items: %w", err)
	}

	merchantMap := make(map[string]*MerchantWithItemsResponse)
	for _, row := range rows {
		merchantID := row.MerchantID.String()

		if _, exists := merchantMap[merchantID]; !exists {
			merchantMap[merchantID] = &MerchantWithItemsResponse{
				Merchant: MerchantInfo{
					MerchantID:       merchantID,
					Name:             row.MerchantName,
					MerchantCategory: row.MerchantCategory,
					ImageUrl:         row.MerchantImageUrl,
					Location: Location{
						Lat:  row.Lat,
						Long: row.Lng,
					},
					CreatedAt: row.MerchantCreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
				},
				Items: []ItemInfo{},
			}
		}

		// if row.ItemID.Valid {
		// 	merchantMap[merchantID].Items = append(merchantMap[merchantID].Items, ItemInfo{
		// 		ItemID:          row.ItemID.UUID.String(),
		// 		Name:            row.ItemName,
		// 		ProductCategory: row.ProductCategory,
		// 		Price:           row.Price,
		// 		ImageUrl:        row.ItemImageUrl,
		// 		CreatedAt:       row.ItemCreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
		// 	})
		// }
	}

	var data []MerchantWithItemsResponse
	for _, m := range merchantMap {
		data = append(data, *m)
	}

	// TODO: Untuk production, pertimbangkan pagination di DB level (lebih kompleks karena grouping)
	// Untuk sekarang, kita asumsikan jumlah merchant terbatas (<100)

	return GetMerchantsNearbyResponse{
		Data: data,
		Meta: PaginationMeta{
			Limit:  0, // bisa diisi jika ada pagination
			Offset: 0,
			Total:  len(data),
		},
	}, nil
}
