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
	"github.com/yenug1k/cars-api/internal/testutil"
)

// ---- helpers -------------------------------------------------------------

func newSvc(t *testing.T) (service.Service, *testutil.MockStore) {
	t.Helper()
	ms := testutil.NewMockStore()
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
	assert.Contains(t, ms.Cars, car.ID)
}

func TestGet_CacheMiss_ThenHit(t *testing.T) {
	svc, ms := newSvc(t)

	seed := &model.Car{ID: "abc", Make: "Honda", Model: "Civic"}
	ms.Cars["abc"] = seed

	c1, err := svc.Get(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, "Honda", c1.Make)

	// Remove from backing store; next call must still return from cache.
	delete(ms.Cars, "abc")

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
		ms.Cars[string(rune('a'+i))] = &model.Car{ID: string(rune('a' + i))}
	}

	resp, err := svc.List(context.Background(), model.ListCarsQuery{PageSize: 5})
	require.NoError(t, err)
	assert.Len(t, resp.Cars, 5)
	assert.NotEmpty(t, resp.NextPageToken)
}

func TestUpdate_Success(t *testing.T) {
	svc, ms := newSvc(t)
	ms.Cars["x"] = &model.Car{ID: "x", Make: "Ford", Status: model.StatusAvailable}

	newStatus := model.StatusSold
	updated, err := svc.Update(context.Background(), "x", &model.UpdateCarRequest{
		Status: &newStatus,
	})

	require.NoError(t, err)
	assert.Equal(t, model.StatusSold, updated.Status)
	assert.Equal(t, "Ford", updated.Make)
}

func TestUpdate_NotFound(t *testing.T) {
	svc, _ := newSvc(t)
	make := "BMW"
	_, err := svc.Update(context.Background(), "missing", &model.UpdateCarRequest{Make: &make})
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func TestDelete_Success(t *testing.T) {
	svc, ms := newSvc(t)
	ms.Cars["del"] = &model.Car{ID: "del"}

	err := svc.Delete(context.Background(), "del")
	require.NoError(t, err)
	assert.NotContains(t, ms.Cars, "del")
}

func TestDelete_NotFound(t *testing.T) {
	svc, _ := newSvc(t)
	err := svc.Delete(context.Background(), "ghost")
	assert.ErrorIs(t, err, store.ErrNotFound)
}
