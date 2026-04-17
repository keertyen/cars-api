package service

import (
	"context"

	"github.com/yenug1k/cars-api/internal/model"
)

type Service interface {
	Create(ctx context.Context, req *model.CreateCarRequest) (*model.Car, error)
	Get(ctx context.Context, id string) (*model.Car, error)
	List(ctx context.Context, q model.ListCarsQuery) (*model.ListCarsResponse, error)
	Update(ctx context.Context, id string, req *model.UpdateCarRequest) (*model.Car, error)
	Delete(ctx context.Context, id string) error
}
