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

type OrderResult struct {
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
		EstimatedDeliveryTimeInMinutes: int(estimatedTime),
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
				Quantity:        int(item.Quantity),
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
	result.EstimatedDeliveryTimeInMinutes = int32(estimate.EstimatedDeliveryTimeInMinutes)

	return result, nil
}

func (r *PurchaseRepository) CreateOrderFromEstimate(ctx context.Context, estimateID uuid.UUID) (OrderResult, error) {
	var result OrderResult

	// Use transaction for performance and consistency
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := r.db.Queries.WithTx(tx)

	// First check if estimate exists
	_, err = txQueries.GetEstimateById(ctx, estimateID)
	if err != nil {
		// This will handle both not found and other database errors
		return result, fmt.Errorf("estimate not found: %w", err)
	}

	// Create the order from the estimate
	order, err := txQueries.CreateOrderFromEstimate(ctx, estimateID)
	if err != nil {
		return result, fmt.Errorf("failed to create order from estimate: %w", err)
	}

	// Get the estimate order details to copy to the order
	estimateDetails, err := txQueries.GetEstimateOrderDetails(ctx, estimateID)
	if err != nil {
		return result, fmt.Errorf("failed to get estimate details: %w", err)
	}

	// Group estimate details by merchant to create order merchants and items properly
	merchantGroups := make(map[string][]database.GetEstimateOrderDetailsRow)
	for _, detail := range estimateDetails {
		merchantStr := detail.MerchantID.String()
		merchantGroups[merchantStr] = append(merchantGroups[merchantStr], detail)
	}

	// Create order merchants and their items
	for _, details := range merchantGroups {
		// All items for the same merchant should belong to the same order merchant record
		// Use the is_starting_point value from the first item for this merchant
		firstDetail := details[0]
		
		orderMerchantID, err := txQueries.CreateOrderMerchant(ctx, database.CreateOrderMerchantParams{
			OrderID:         order.ID,
			MerchantID:      firstDetail.MerchantID,
			IsStartingPoint: firstDetail.IsStartingPoint,
		})
		if err != nil {
			return result, fmt.Errorf("failed to create order merchant: %w", err)
		}

		// Create order items for this merchant
		for _, detail := range details {
			err = txQueries.CreateOrderItem(ctx, database.CreateOrderItemParams{
				OrderMerchantID: orderMerchantID,
				ItemID:          detail.ItemID,
				Quantity:        detail.Quantity,
			})
			if err != nil {
				return result, fmt.Errorf("failed to create order item: %w", err)
			}
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result.ID = order.ID
	result.TotalPrice = order.TotalPrice
	result.EstimatedDeliveryTimeInMinutes = int32(order.EstimatedDeliveryTimeInMinutes)

	return result, nil
}

// Helper method to get an estimate by ID (for validation)
func (r *PurchaseRepository) GetEstimateById(ctx context.Context, estimateID uuid.UUID) (database.Estimates, error) {
	return r.db.Queries.GetEstimateById(ctx, estimateID)
}
