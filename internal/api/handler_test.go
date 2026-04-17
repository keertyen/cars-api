package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenug1k/cars-api/config"
	"github.com/yenug1k/cars-api/internal/api"
	"github.com/yenug1k/cars-api/internal/model"
	"github.com/yenug1k/cars-api/internal/store"
)

// ---- mock service --------------------------------------------------------

type mockService struct {
	cars map[string]*model.Car
}

func newMockService() *mockService {
	return &mockService{cars: make(map[string]*model.Car)}
}

func (m *mockService) Create(_ context.Context, req *model.CreateCarRequest) (*model.Car, error) {
	car := &model.Car{ID: "new-id", Make: req.Make, Model: req.Model, Status: req.Status}
	m.cars["new-id"] = car
	return car, nil
}
func (m *mockService) Get(_ context.Context, id string) (*model.Car, error) {
	c, ok := m.cars[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return c, nil
}
func (m *mockService) List(_ context.Context, _ model.ListCarsQuery) (*model.ListCarsResponse, error) {
	cars := make([]*model.Car, 0, len(m.cars))
	for _, c := range m.cars {
		cars = append(cars, c)
	}
	return &model.ListCarsResponse{Cars: cars}, nil
}
func (m *mockService) Update(_ context.Context, id string, req *model.UpdateCarRequest) (*model.Car, error) {
	c, ok := m.cars[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	if req.Make != nil {
		c.Make = *req.Make
	}
	return c, nil
}
func (m *mockService) Delete(_ context.Context, id string) error {
	if _, ok := m.cars[id]; !ok {
		return store.ErrNotFound
	}
	delete(m.cars, id)
	return nil
}

// ---- helpers -------------------------------------------------------------

func newTestApp(svc *mockService) *api.Handler {
	return api.NewHandler(svc)
}

// testCfg returns a config with a high rate limit so tests are never throttled.
func testCfg() *config.Config {
	return &config.Config{
		Port:            "3000",
		RateLimitRPS:    1000,
		RateLimitBurst:  1000,
		CacheTTLSeconds: 300,
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func do(t *testing.T, svc *mockService, method, path string, body []byte) *http.Response {
	t.Helper()
	app := api.NewApp(svc, testLogger(), testCfg())
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	return resp
}

func readJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, v))
}

// ---- tests ---------------------------------------------------------------

func TestCreateCar_Success(t *testing.T) {
	svc := newMockService()
	body := []byte(`{"make":"Toyota","model":"Camry","year":2022,"color":"Blue","price":25000,"status":"available"}`)
	resp := do(t, svc, http.MethodPost, "/v1/cars", body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var car model.Car
	readJSON(t, resp, &car)
	assert.Equal(t, "Toyota", car.Make)
}

func TestCreateCar_ValidationError(t *testing.T) {
	svc := newMockService()
	resp := do(t, svc, http.MethodPost, "/v1/cars", []byte(`{"make":"Toyota"}`))

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var er map[string]any
	readJSON(t, resp, &er)
	assert.Equal(t, "validation failed", er["error"])
}

func TestGetCar_Success(t *testing.T) {
	svc := newMockService()
	svc.cars["car-1"] = &model.Car{ID: "car-1", Make: "Honda"}
	resp := do(t, svc, http.MethodGet, "/v1/cars/car-1", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var car model.Car
	readJSON(t, resp, &car)
	assert.Equal(t, "Honda", car.Make)
}

func TestGetCar_NotFound(t *testing.T) {
	svc := newMockService()
	resp := do(t, svc, http.MethodGet, "/v1/cars/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListCars(t *testing.T) {
	svc := newMockService()
	svc.cars["a"] = &model.Car{ID: "a", Make: "BMW"}
	svc.cars["b"] = &model.Car{ID: "b", Make: "Audi"}
	resp := do(t, svc, http.MethodGet, "/v1/cars?page_size=10", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var result model.ListCarsResponse
	readJSON(t, resp, &result)
	assert.Len(t, result.Cars, 2)
}

func TestListCars_InvalidPageSize(t *testing.T) {
	svc := newMockService()
	resp := do(t, svc, http.MethodGet, "/v1/cars?page_size=999", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateCar_Success(t *testing.T) {
	svc := newMockService()
	svc.cars["u1"] = &model.Car{ID: "u1", Make: "Ford"}
	resp := do(t, svc, http.MethodPut, "/v1/cars/u1", []byte(`{"make":"Tesla"}`))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var car model.Car
	readJSON(t, resp, &car)
	assert.Equal(t, "Tesla", car.Make)
}

func TestUpdateCar_NotFound(t *testing.T) {
	svc := newMockService()
	resp := do(t, svc, http.MethodPut, "/v1/cars/ghost", []byte(`{"make":"X"}`))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteCar_Success(t *testing.T) {
	svc := newMockService()
	svc.cars["d1"] = &model.Car{ID: "d1"}
	resp := do(t, svc, http.MethodDelete, "/v1/cars/d1", nil)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteCar_NotFound(t *testing.T) {
	svc := newMockService()
	resp := do(t, svc, http.MethodDelete, "/v1/cars/phantom", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHealth(t *testing.T) {
	svc := newMockService()
	resp := do(t, svc, http.MethodGet, "/health", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
