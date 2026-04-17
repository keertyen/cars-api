package store

import (
	"context"
	"errors"

	"github.com/yenug1k/cars-api/internal/model"
)

var ErrNotFound = errors.New("car not found")

type Store interface {
	Create(ctx context.Context, car *model.Car) error
	Get(ctx context.Context, id string) (*model.Car, error)
	List(ctx context.Context, q model.ListCarsQuery) ([]*model.Car, error)
	Update(ctx context.Context, car *model.Car) error
	Delete(ctx context.Context, id string) error
}
