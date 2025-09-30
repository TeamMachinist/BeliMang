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
// - nil → all merchants definitely ≤3000m
// - ErrNeedExactValidation → some in ambiguous zone (2500–3500m)
// - error → definitely >3000m
func (s *PurchaseService) validateWithH3PreFilter(
	userH3 h3.Cell,
	merchantPoints []merchantPoint,
) error {
	var ambiguous bool
	for _, mp := range merchantPoints {
		dist := utils.H3GridDistanceMeters(userH3, mp.H3Cell)
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

	// === Ambil data & konversi ke H3 ===
	var points []merchantPoint
	totalPrice := int64(0)

	userH3, err := utils.LatLonToH3(req.UserLocation.Lat, req.UserLocation.Long)
	if err != nil {
		return EstimateResponse{}, errors.New("invalid user location")
	}

	for _, o := range req.Orders {
		parsedMerchantID, _ := uuid.Parse(o.MerchantID)
		merchant, err := s.queries.GetMerchantLatLong(ctx, parsedMerchantID)
		if err != nil {
			return EstimateResponse{}, errors.New("merchant not found")
		}

		h3Cell, err := utils.LatLonToH3(merchant.Lat, merchant.Lng)
		if err != nil {
			return EstimateResponse{}, errors.New("invalid merchant location")
		}

		for _, item := range o.Items {
			parsedItemID, _ := uuid.Parse(item.ItemID)
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
			H3Cell:     h3Cell,
			IsStart:    o.IsStartingPoint,
			Order:      o,
		})
	}

	// === Validasi jarak: H3 pre-filter ===
	err = s.validateWithH3PreFilter(userH3, points)
	if err != nil && !errors.Is(err, ErrNeedExactValidation) {
		return EstimateResponse{}, errors.New("coordinates too far")
	}

	var useHaversineForTSP bool
	if errors.Is(err, ErrNeedExactValidation) {
		// Fallback: validasi exact dengan Haversine
		for _, mp := range points {
			d := utils.HaversineDistance(
				req.UserLocation.Lat, req.UserLocation.Long,
				mp.Lat, mp.Lng,
			)
			if d > 3000 {
				return EstimateResponse{}, errors.New("coordinates too far")
			}
		}
		useHaversineForTSP = false // use harversine for TSP as well
	}

	// === Bangun rute TSP ===
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
			// Gunakan H3 GridDistance untuk TSP
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

	// === Estimasi waktu ===
	timeMinutes := utils.EstimateTimeMinutes(totalDist)

	// === Simpan estimate dengan repository ===
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
