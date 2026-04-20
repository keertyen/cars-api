package testutil

import (
	"context"

	"github.com/yenug1k/cars-api/internal/model"
	"github.com/yenug1k/cars-api/internal/store"
)

type MockStore struct {
	Cars map[string]*model.Car
}

func NewMockStore() *MockStore { return &MockStore{Cars: make(map[string]*model.Car)} }

func (m *MockStore) Create(_ context.Context, car *model.Car) error {
	m.Cars[car.ID] = car
	return nil
}

func (m *MockStore) Get(_ context.Context, id string) (*model.Car, error) {
	c, ok := m.Cars[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return c, nil
}

func (m *MockStore) List(_ context.Context, _ model.ListCarsQuery) ([]*model.Car, error) {
	out := make([]*model.Car, 0, len(m.Cars))
	for _, c := range m.Cars {
		out = append(out, c)
	}
	return out, nil
}

func (m *MockStore) Update(_ context.Context, car *model.Car) error {
	if _, ok := m.Cars[car.ID]; !ok {
		return store.ErrNotFound
	}
	m.Cars[car.ID] = car
	return nil
}

func (m *MockStore) Delete(_ context.Context, id string) error {
	delete(m.Cars, id)
	return nil
}
