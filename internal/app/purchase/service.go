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

var ErrNeedExactValidation = errors.New("ambiguous distance: need exact validation")

type PurchaseService struct {
	queries *database.Queries
}

func NewPurchaseService(q *database.Queries) *PurchaseService {
	return &PurchaseService{queries: q}
}

// validateWithH3PreFilter returns:
// - nil → all merchants definitely ≤3000m
// - ErrNeedExactValidation → some in ambiguous zone (2500–3500m)
// - error → definitely >3000m or invalid
func (s *PurchaseService) validateWithH3PreFilter(
	userLat, userLng float64,
	merchantPoints []merchantPoint,
) error {
	userH3, err := utils.LatLonToH3(userLat, userLng)
	if err != nil {
		return ErrNeedExactValidation
	}

	var ambiguous bool
	for _, mp := range merchantPoints {
		mH3, err := utils.LatLonToH3(mp.Lat, mp.Lng)
		// fmt.Println("Merchant", mp.MerchantID, "H3:", mH3)
		if err != nil {
			return ErrNeedExactValidation
		}
		dist := utils.H3GridDistanceMeters(userH3, mH3)
		if dist < 0 {
			return ErrNeedExactValidation
		}
		if dist > utils.SAFE_REJECT_M {
			return errors.New("coordinates too far")
		}
		if dist >= utils.SAFE_ACCEPT_M {
			ambiguous = true
		}
	}
	if ambiguous {
		return ErrNeedExactValidation
	}
	return nil
}

type merchantPoint struct {
	MerchantID string
	Lat, Lng   float64
	IsStart    bool
	Order      Order
}

func (s *PurchaseService) ValidateAndEstimate(ctx context.Context, req EstimateRequest) (EstimateResponse, error) {
	// === Validasi request ===
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

	// === Ambil data merchant & item ===
	var points []merchantPoint
	totalPrice := int64(0)

	for _, o := range req.Orders {
		parsedMerchantID, _ := uuid.Parse(o.MerchantID)

		merchant, err := s.queries.GetMerchantLatLong(ctx, parsedMerchantID)
		if err != nil {
			return EstimateResponse{}, errors.New("merchant not found")
		}

		for _, item := range o.Items {

			parsedItemID, _ := uuid.Parse(item.ItemID)
			parsedMerchantID, _ := uuid.Parse(o.MerchantID)
			price, err := s.queries.GetItemPrice(ctx, database.GetItemPriceParams{
				ItemID:     parsedItemID,
				MerchantID: parsedMerchantID,
			})
			if err != nil {
				return EstimateResponse{}, errors.New("item not found")
			}
			totalPrice += price * item.Quantity
		}

		points = append(points, merchantPoint{
			MerchantID: o.MerchantID,
			Lat:        merchant.Lat,
			Lng:        merchant.Lng,
			IsStart:    o.IsStartingPoint,
			Order:      o,
		})
	}

	// === Validasi jarak via H3 pre-filter ===
	err := s.validateWithH3PreFilter(req.UserLocation.Lat, req.UserLocation.Long, points)
	if err != nil && !errors.Is(err, ErrNeedExactValidation) {
		return EstimateResponse{}, errors.New("coordinates too far")
	}

	// === Jika ambigu, validasi dengan Haversine (exact) ===
	if errors.Is(err, ErrNeedExactValidation) {
		for _, mp := range points {
			d := utils.HaversineDistance(
				req.UserLocation.Lat, req.UserLocation.Long,
				mp.Lat, mp.Lng,
			)
			if d > 3000 {
				return EstimateResponse{}, errors.New("coordinates too far")
			}
		}
	}

	// === Bangun rute Nearest Neighbor ===
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

	// === Hitung total jarak ===
	totalDist := 0.0
	for i := 0; i < len(route)-1; i++ {
		totalDist += utils.HaversineDistance(route[i].Lat, route[i].Lng, route[i+1].Lat, route[i+1].Lng)
	}
	last := route[len(route)-1]
	totalDist += utils.HaversineDistance(last.Lat, last.Lng, req.UserLocation.Lat, req.UserLocation.Long)

	// === Estimasi waktu ===
	timeMinutes := utils.EstimateTimeMinutes(totalDist)

	orderJson, err := json.Marshal(req.Orders)
	if err != nil {
		return EstimateResponse{}, fmt.Errorf("failed to marshal orders: %w", err)
	}

	// === Simpan estimate ===
	estimate, err := s.queries.CreateEstimate(ctx, database.CreateEstimateParams{
		UserLat:                        req.UserLocation.Lat,
		UserLng:                        req.UserLocation.Long,
		Orders:                         orderJson,
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
