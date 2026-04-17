package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenug1k/cars-api/internal/cache"
	"github.com/yenug1k/cars-api/internal/model"
	"github.com/yenug1k/cars-api/internal/service"
	"github.com/yenug1k/cars-api/internal/store"
)

// ---- in-memory mock store -----------------------------------------------

type mockStore struct {
	cars map[string]*model.Car
}

func newMockStore() *mockStore { return &mockStore{cars: make(map[string]*model.Car)} }

func (m *mockStore) Create(_ context.Context, car *model.Car) error {
	m.cars[car.ID] = car
	return nil
}
func (m *mockStore) Get(_ context.Context, id string) (*model.Car, error) {
	c, ok := m.cars[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return c, nil
}
func (m *mockStore) List(_ context.Context, _ model.ListCarsQuery) ([]*model.Car, error) {
	out := make([]*model.Car, 0, len(m.cars))
	for _, c := range m.cars {
		out = append(out, c)
	}
	return out, nil
}
func (m *mockStore) Update(_ context.Context, car *model.Car) error {
	if _, ok := m.cars[car.ID]; !ok {
		return store.ErrNotFound
	}
	m.cars[car.ID] = car
	return nil
}
func (m *mockStore) Delete(_ context.Context, id string) error {
	delete(m.cars, id)
	return nil
}

// ---- helpers -------------------------------------------------------------

func newSvc(t *testing.T) (service.Service, *mockStore) {
	t.Helper()
	ms := newMockStore()
	c := cache.New(5*time.Minute, 10*time.Minute)
	t.Cleanup(c.Close)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.New(ms, c, logger), ms
}

func validCreateReq() *model.CreateCarRequest {
	return &model.CreateCarRequest{
		Make:    "Toyota",
		Model:   "Camry",
		Year:    2022,
		Color:   "Blue",
		Price:   25000.00,
		Mileage: 0,
		Status:  model.StatusAvailable,
	}
}

// ---- tests ---------------------------------------------------------------

func TestCreate_Success(t *testing.T) {
	svc, ms := newSvc(t)

	car, err := svc.Create(context.Background(), validCreateReq())

	require.NoError(t, err)
	assert.NotEmpty(t, car.ID)
	assert.Equal(t, "Toyota", car.Make)
	assert.Equal(t, model.StatusAvailable, car.Status)
	assert.False(t, car.CreatedAt.IsZero())
	// must be persisted
	assert.Contains(t, ms.cars, car.ID)
}

func TestGet_CacheMiss_ThenHit(t *testing.T) {
	svc, ms := newSvc(t)

	seed := &model.Car{ID: "abc", Make: "Honda", Model: "Civic"}
	ms.cars["abc"] = seed

	c1, err := svc.Get(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, "Honda", c1.Make)

	// Remove from backing store; next call must still return from cache.
	delete(ms.cars, "abc")

	c2, err := svc.Get(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, c1.Make, c2.Make)
}

func TestGet_NotFound(t *testing.T) {
	svc, _ := newSvc(t)
	_, err := svc.Get(context.Background(), "does-not-exist")
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func TestList_Pagination(t *testing.T) {
	svc, ms := newSvc(t)
	for i := 0; i < 5; i++ {
		ms.cars[string(rune('a'+i))] = &model.Car{ID: string(rune('a' + i))}
	}

	resp, err := svc.List(context.Background(), model.ListCarsQuery{PageSize: 5})
	require.NoError(t, err)
	assert.Len(t, resp.Cars, 5)
	// exactly PageSize results → nextToken should be set
	assert.NotEmpty(t, resp.NextPageToken)
}

func TestUpdate_Success(t *testing.T) {
	svc, ms := newSvc(t)
	ms.cars["x"] = &model.Car{ID: "x", Make: "Ford", Status: model.StatusAvailable}

	newStatus := model.StatusSold
	updated, err := svc.Update(context.Background(), "x", &model.UpdateCarRequest{
		Status: &newStatus,
	})

	require.NoError(t, err)
	assert.Equal(t, model.StatusSold, updated.Status)
	assert.Equal(t, "Ford", updated.Make) // untouched field preserved
}

func TestUpdate_NotFound(t *testing.T) {
	svc, _ := newSvc(t)
	make := "BMW"
	_, err := svc.Update(context.Background(), "missing", &model.UpdateCarRequest{Make: &make})
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func TestDelete_Success(t *testing.T) {
	svc, ms := newSvc(t)
	ms.cars["del"] = &model.Car{ID: "del"}

	err := svc.Delete(context.Background(), "del")
	require.NoError(t, err)
	assert.NotContains(t, ms.cars, "del")
}

func TestDelete_NotFound(t *testing.T) {
	svc, _ := newSvc(t)
	err := svc.Delete(context.Background(), "ghost")
	assert.ErrorIs(t, err, store.ErrNotFound)
}
