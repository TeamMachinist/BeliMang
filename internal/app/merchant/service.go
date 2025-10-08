package merchant

import (
	"context"
	"fmt"
	"sync"
	"time"

	"belimang/internal/infrastructure/cache"
	"belimang/internal/infrastructure/database"
	logger "belimang/internal/pkg/logging"

	"github.com/google/uuid"
)

// MerchantService handles Merchant business logic
type MerchantService struct {
	cache *cache.RedisCache
	db    database.Querier
}

// NewMerchantService creates a new MerchantService
func NewMerchantService(cache *cache.RedisCache, database database.Querier) *MerchantService {
	return &MerchantService{cache: cache, db: database}
}

func (s *MerchantService) CreateMerchantService(ctx context.Context, adminID uuid.UUID, req PostMerchantRequest) (PostMerchantResponse, error) {
	logger.InfoCtx(ctx, "Create new merchant process", "id", adminID, "request", req)
	// Should check cache Admin existed?

	// Should check db admin existed?

	// update db merchant
	rows, err := s.db.CreateMerchant(ctx, database.CreateMerchantParams{
		AdminID:          adminID,
		Name:             req.Name,
		MerchantCategory: req.MerchantCategory,
		ImageUrl:         req.ImageURL,
		Lat:              req.Location.Latitude,
		Lng:              req.Location.Longitude,
	})
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to create merchant", "error", err)
		return PostMerchantResponse{}, err
	}

	resp := PostMerchantResponse{
		MerchantID: rows.ID.String(),
	}

	// set cache merchant
	s.cache.Set(ctx, fmt.Sprintf(cache.MerchantKey, rows.ID), resp, cache.MerchantTTL)
	s.cache.Exists(ctx, fmt.Sprintf(cache.MerchantExistsKey, rows.ID))

	logger.InfoCtx(ctx, "Merchant created successfully", "resp", resp, "rows", rows)
	return resp, nil
}

// SearchMerchantsService searches merchants using filter params
func (s *MerchantService) SearchMerchantsService(ctx context.Context, filter MerchantFilter) (GetMerchantsResponse, error) {
	logger.InfoCtx(ctx, "Create search merchants process", "merchantId", filter.MerchantID, "name", filter.Name, "category", filter.MerchantCategory, "sort", filter.CreatedAtSort)

	var merchantId uuid.UUID
	if filter.MerchantID != "" {
		id, err := uuid.Parse(filter.MerchantID)
		if err == nil {
			merchantId = id
		}
	}

	merchantCategory := filter.MerchantCategory
	if merchantCategory != "" {
		if _, ok := validMerchantCategories[merchantCategory]; !ok {
			return GetMerchantsResponse{
				Data: []Merchant{},
				Meta: Meta{Limit: filter.Limit, Offset: filter.Offset, Total: 0},
			}, nil
		}
	}

	limit := filter.Limit
	offset := filter.Offset
	if limit <= 0 {
		limit = 5
	}
	if offset < 0 {
		offset = 0
	}

	logger.DebugCtx(ctx, "Query search merchants", "merchantId", merchantId, "name", filter.Name, "category", merchantCategory, "sort", filter.CreatedAtSort)

	// Execute both queries in parallel for maximum performance
	var (
		data     []Merchant
		total    int64
		errItems error
		errCount error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	// Fetch items (uses index scan with LIMIT - fast!)
	go func() {
		defer wg.Done()
		if filter.CreatedAtSort == "asc" {
			rows, err := s.db.SearchMerchantsAsc(ctx, database.SearchMerchantsAscParams{
				Column1: merchantId,
				Column2: filter.Name,
				Column3: filter.MerchantCategory,
				Limit:   int32(limit),
				Offset:  int32(offset),
			})

			if err != nil {
				errItems = err
				return
			}

			logger.InfoCtx(ctx, "SearchMerchantsAsc queried successfully ", "rows", rows)

			// Convert immediately
			data = make([]Merchant, len(rows))
			for i, row := range rows {
				data[i] = Merchant{
					MerchantID:       row.ID.String(),
					Name:             row.Name,
					MerchantCategory: row.MerchantCategory,
					ImageURL:         row.ImageUrl,
					Location:         Location{Latitude: row.Lat, Longitude: row.Lng},
					CreatedAt:        row.CreatedAt.Format(time.RFC3339Nano),
				}
			}
		} else {
			rows, err := s.db.SearchMerchantsDesc(ctx, database.SearchMerchantsDescParams{
				Column1: merchantId,
				Column2: filter.Name,
				Column3: filter.MerchantCategory,
				Limit:   int32(limit),
				Offset:  int32(offset),
			})

			if err != nil {
				errItems = err
				return
			}

			logger.InfoCtx(ctx, "SearchMerchantsDesc queried successfully ", "rows", rows)

			// Convert immediately
			data = make([]Merchant, len(rows))
			for i, row := range rows {
				data[i] = Merchant{
					MerchantID:       row.ID.String(),
					Name:             row.Name,
					MerchantCategory: row.MerchantCategory,
					ImageURL:         row.ImageUrl,
					Location:         Location{Latitude: row.Lat, Longitude: row.Lng},
					CreatedAt:        row.CreatedAt.Format(time.RFC3339Nano),
				}
			}
		}
	}()

	// Fetch count
	go func() {
		defer wg.Done()
		total, errCount = s.db.CountSearchMerchants(ctx, database.CountSearchMerchantsParams{
			Column1: merchantId,
			Column2: filter.Name,
			Column3: filter.MerchantCategory,
		})
	}()

	wg.Wait()

	if errItems != nil {
		return GetMerchantsResponse{}, errItems
	}
	if errCount != nil {
		return GetMerchantsResponse{}, errCount
	}

	meta := Meta{Limit: limit, Offset: offset, Total: int(total)}
	logger.InfoCtx(ctx, "Merchant searched successfully", "data", data, "Meta", meta)
	return GetMerchantsResponse{Data: data, Meta: meta}, nil
}
