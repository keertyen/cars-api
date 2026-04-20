package testutil

import (
	"context"
	"fmt"

	"github.com/yenug1k/cars-api/internal/model"
	"github.com/yenug1k/cars-api/internal/store"
)

type MockService struct {
	Cars map[string]*model.Car
	seq  int
}

func NewMockService() *MockService { return &MockService{Cars: make(map[string]*model.Car)} }

func (m *MockService) Create(_ context.Context, req *model.CreateCarRequest) (*model.Car, error) {
	m.seq++
	id := fmt.Sprintf("id-%d", m.seq)
	car := &model.Car{ID: id, Make: req.Make, Model: req.Model, Status: req.Status}
	m.Cars[id] = car
	return car, nil
}

func (m *MockService) Get(_ context.Context, id string) (*model.Car, error) {
	c, ok := m.Cars[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return c, nil
}

func (m *MockService) List(_ context.Context, _ model.ListCarsQuery) (*model.ListCarsResponse, error) {
	cars := make([]*model.Car, 0, len(m.Cars))
	for _, c := range m.Cars {
		cars = append(cars, c)
	}
	return &model.ListCarsResponse{Cars: cars}, nil
}

func (m *MockService) Update(_ context.Context, id string, req *model.UpdateCarRequest) (*model.Car, error) {
	c, ok := m.Cars[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	if req.Make != nil {
		c.Make = *req.Make
	}
	return c, nil
}

func (m *MockService) Delete(_ context.Context, id string) error {
	if _, ok := m.Cars[id]; !ok {
		return store.ErrNotFound
	}
	delete(m.Cars, id)
	return nil
}
