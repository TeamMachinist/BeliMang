package purchase

import (
	"belimang/internal/infrastructure/database"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type PurchaseRepository struct {
	db *database.DB
}

func NewPurchaseRepository(db *database.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

type EstimateResult struct {
	ID                             uuid.UUID
	TotalPrice                     int64
	EstimatedDeliveryTimeInMinutes int32
}

func (r *PurchaseRepository) CreateEstimateWithOrders(ctx context.Context, userLat, userLng, totalPrice float64, estimatedTime int32, orders []Order) (EstimateResult, error) {
	var result EstimateResult

	// Use transaction for performance and consistency
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := r.db.Queries.WithTx(tx)

	// Create the estimate
	estimate, err := txQueries.CreateEstimate(ctx, database.CreateEstimateParams{
		UserLat:                        userLat,
		UserLng:                        userLng,
		TotalPrice:                     int64(totalPrice),
		EstimatedDeliveryTimeInMinutes: int32(estimatedTime),
	})
	if err != nil {
		return result, fmt.Errorf("failed to save estimate: %w", err)
	}

	// Create estimate orders in batch
	for _, order := range orders {
		parsedMerchantID, err := uuid.Parse(order.MerchantID)
		if err != nil {
			return result, fmt.Errorf("invalid merchant id: %w", err)
		}

		err = txQueries.CreateEstimateOrder(ctx, database.CreateEstimateOrderParams{
			EstimateID:      estimate.ID,
			MerchantID:      parsedMerchantID,
			IsStartingPoint: order.IsStartingPoint,
		})
		if err != nil {
			return result, fmt.Errorf("failed to save estimate order: %w", err)
		}
	}

	// Get the created estimate order IDs to link with items
	estimateOrders, err := txQueries.GetEstimateOrderIds(ctx, estimate.ID)
	if err != nil {
		return result, fmt.Errorf("failed to get estimate orders: %w", err)
	}

	// Create a mapping from merchant_id to estimate_order_id for quick lookup
	orderIdMap := make(map[uuid.UUID]uuid.UUID)
	for _, eo := range estimateOrders {
		orderIdMap[eo.MerchantID] = eo.ID
	}

	// Create all items in batch
	for _, order := range orders {
		parsedMerchantID, err := uuid.Parse(order.MerchantID)
		if err != nil {
			return result, fmt.Errorf("invalid merchant id: %w", err)
		}

		estimateOrderId, exists := orderIdMap[parsedMerchantID]
		if !exists {
			return result, fmt.Errorf("estimate order not found for merchant: %s", order.MerchantID)
		}

		for _, item := range order.Items {
			parsedItemID, err := uuid.Parse(item.ItemID)
			if err != nil {
				return result, fmt.Errorf("invalid item id: %w", err)
			}

			err = txQueries.CreateEstimateOrderItem(ctx, database.CreateEstimateOrderItemParams{
				EstimateOrderID: estimateOrderId,
				ItemID:          parsedItemID,
				Quantity:        int32(item.Quantity),
			})
			if err != nil {
				return result, fmt.Errorf("failed to save estimate order item: %w", err)
			}
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result.ID = estimate.ID
	result.TotalPrice = estimate.TotalPrice
	result.EstimatedDeliveryTimeInMinutes = estimate.EstimatedDeliveryTimeInMinutes

	return result, nil
}
