package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/yenug1k/cars-api/internal/cache"
	"github.com/yenug1k/cars-api/internal/model"
	"github.com/yenug1k/cars-api/internal/store"
)

const (
	cacheCarPrefix  = "car:"
	cacheListPrefix = "cars:list:"
	listCacheTTL    = 30 * time.Second
)

type CarService struct {
	store  store.Store
	cache  *cache.Cache
	logger *slog.Logger
}

func New(s store.Store, c *cache.Cache, logger *slog.Logger) Service {
	return &CarService{store: s, cache: c, logger: logger}
}

func (s *CarService) Create(ctx context.Context, req *model.CreateCarRequest) (*model.Car, error) {
	now := time.Now().UTC()
	car := &model.Car{
		ID:        uuid.New().String(),
		Make:      req.Make,
		Model:     req.Model,
		Year:      req.Year,
		Color:     req.Color,
		Price:     req.Price,
		Mileage:   req.Mileage,
		VIN:       req.VIN,
		Status:    req.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.Create(ctx, car); err != nil {
		return nil, fmt.Errorf("service.Create: %w", err)
	}

	s.cache.Set(cacheCarPrefix+car.ID, car)
	s.cache.DeletePrefix(cacheListPrefix)

	s.logger.Info("car created", "id", car.ID, "make", car.Make, "model", car.Model)
	return car, nil
}

func (s *CarService) Get(ctx context.Context, id string) (*model.Car, error) {
	key := cacheCarPrefix + id
	if v, ok := s.cache.Get(key); ok {
		return v.(*model.Car), nil
	}

	car, err := s.store.Get(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("service.Get: %w", err)
	}

	s.cache.Set(key, car)
	return car, nil
}

func (s *CarService) List(ctx context.Context, q model.ListCarsQuery) (*model.ListCarsResponse, error) {
	key := fmt.Sprintf("%s%d:%s", cacheListPrefix, q.PageSize, q.PageToken)
	if v, ok := s.cache.Get(key); ok {
		return v.(*model.ListCarsResponse), nil
	}

	cars, err := s.store.List(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("service.List: %w", err)
	}

	var nextToken string
	if len(cars) == q.PageSize {
		nextToken = cars[len(cars)-1].ID
	}

	resp := &model.ListCarsResponse{Cars: cars, NextPageToken: nextToken}
	s.cache.SetTTL(key, resp, listCacheTTL)
	return resp, nil
}

func (s *CarService) Update(ctx context.Context, id string, req *model.UpdateCarRequest) (*model.Car, error) {
	car, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Make != nil {
		car.Make = *req.Make
	}
	if req.Model != nil {
		car.Model = *req.Model
	}
	if req.Year != nil {
		car.Year = *req.Year
	}
	if req.Color != nil {
		car.Color = *req.Color
	}
	if req.Price != nil {
		car.Price = *req.Price
	}
	if req.Mileage != nil {
		car.Mileage = *req.Mileage
	}
	if req.VIN != nil {
		car.VIN = *req.VIN
	}
	if req.Status != nil {
		car.Status = *req.Status
	}
	car.UpdatedAt = time.Now().UTC()

	if err := s.store.Update(ctx, car); err != nil {
		return nil, fmt.Errorf("service.Update: %w", err)
	}

	s.cache.Set(cacheCarPrefix+id, car)
	s.cache.DeletePrefix(cacheListPrefix)

	s.logger.Info("car updated", "id", id)
	return car, nil
}

func (s *CarService) Delete(ctx context.Context, id string) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("service.Delete: %w", err)
	}

	s.cache.Delete(cacheCarPrefix + id)
	s.cache.DeletePrefix(cacheListPrefix)

	s.logger.Info("car deleted", "id", id)
	return nil
}
