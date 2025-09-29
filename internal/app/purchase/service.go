package purchase

import (
	"belimang/internal/infrastructure/database"
	"belimang/internal/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type PurchaseService struct {
	queries *database.Queries
}

func NewPurchaseService(q *database.Queries) *PurchaseService {
	return &PurchaseService{queries: q}
}

func (s *PurchaseService) ValidateAndEstimate(ctx context.Context, req EstimateRequest) (EstimateResponse, error) {
	// === Validasi struktur request ===
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

	// === Ekstrak merchant IDs ===
	merchantIDs := make([]uuid.UUID, len(req.Orders))
	for i, o := range req.Orders {
		id, err := uuid.Parse(o.MerchantID)
		if err != nil {
			return EstimateResponse{}, errors.New("invalid merchantId")
		}
		merchantIDs[i] = id
	}

	// === Validasi jarak via PostGIS (≤3000m) ===
	tooFar, err := s.queries.ValidateMerchantsWithin3km(ctx, database.ValidateMerchantsWithin3kmParams{
		MerchantIds: merchantIDs,
		UserLng:     req.UserLocation.Long,
		UserLat:     req.UserLocation.Lat,
	})
	if err != nil {
		return EstimateResponse{}, fmt.Errorf("distance validation failed: %w", err)
	}
	if len(tooFar) > 0 {
		return EstimateResponse{}, errors.New("coordinates too far")
	}

	// === Ambil data merchant & hitung total harga ===
	type point struct {
		MerchantID string
		Lat, Lng   float64
		IsStart    bool
	}

	var points []point
	totalPrice := int64(0)

	for _, o := range req.Orders {
		parsedID, _ := uuid.Parse(o.MerchantID)

		merchant, err := s.queries.GetMerchantLatLong(ctx, parsedID)
		if err != nil {
			return EstimateResponse{}, errors.New("merchant not found")
		}

		for _, item := range o.Items {
			parsedID, _ := uuid.Parse(item.ItemID)
			parsedMerchantID, _ := uuid.Parse(o.MerchantID)

			price, err := s.queries.GetItemPrice(ctx, database.GetItemPriceParams{
				ItemID:     parsedID,
				MerchantID: parsedMerchantID,
			})
			if err != nil {
				return EstimateResponse{}, errors.New("item not found")
			}
			totalPrice += price * item.Quantity
		}

		points = append(points, point{
			MerchantID: o.MerchantID,
			Lat:        merchant.Lat,
			Lng:        merchant.Lng,
			IsStart:    o.IsStartingPoint,
		})
	}

	// === Bangun rute Nearest Neighbor TSP ===
	var startPoint *point
	rest := make([]point, 0)
	for i := range points {
		if points[i].IsStart {
			startPoint = &points[i]
		} else {
			rest = append(rest, points[i])
		}
	}
	if startPoint == nil {
		return EstimateResponse{}, errors.New("starting point not found")
	}

	route := []point{*startPoint}
	current := *startPoint

	for len(rest) > 0 {
		bestIdx := -1
		bestDist := 1e15

		for i, p := range rest {
			d := utils.HaversineDistance(current.Lat, current.Lng, p.Lat, p.Lng)
			if d < bestDist {
				bestDist = d
				bestIdx = i
			}
		}

		if bestIdx == -1 {
			break
		}

		route = append(route, rest[bestIdx])
		current = rest[bestIdx]
		rest = append(rest[:bestIdx], rest[bestIdx+1:]...)
	}

	// === Hitung total jarak: merchant → merchant → user ===
	totalDist := 0.0
	for i := 0; i < len(route)-1; i++ {
		totalDist += utils.HaversineDistance(route[i].Lat, route[i].Lng, route[i+1].Lat, route[i+1].Lng)
	}
	// Terakhir ke user
	last := route[len(route)-1]
	totalDist += utils.HaversineDistance(last.Lat, last.Lng, req.UserLocation.Lat, req.UserLocation.Long)

	// === Estimasi waktu (menit) ===
	timeMinutes := utils.EstimateTimeMinutes(totalDist)
	ordersJSON, err := json.Marshal(req)
	if err != nil {
		return EstimateResponse{}, fmt.Errorf("failed to marshal orders: %w", err)
	}

	// === Simpan ke database ===
	estimate, err := s.queries.CreateEstimate(ctx, database.CreateEstimateParams{
		UserLng:                        req.UserLocation.Long,
		UserLat:                        req.UserLocation.Lat,
		Orders:                         ordersJSON,
		TotalPrice:                     totalPrice,
		EstimatedDeliveryTimeInMinutes: int32(timeMinutes),
	})
	if err != nil {
		return EstimateResponse{}, fmt.Errorf("failed to save estimate: %w", err)
	}

	return EstimateResponse{
		TotalPrice:                     estimate.TotalPrice,
		EstimatedDeliveryTimeInMinutes: int(estimate.EstimatedDeliveryTimeInMinutes),
		CalculatedEstimateId:           estimate.ID.String(),
	}, nil
}
