package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenug1k/cars-api/config"
	"github.com/yenug1k/cars-api/internal/api"
	"github.com/yenug1k/cars-api/internal/model"
	"github.com/yenug1k/cars-api/internal/testutil"
)

// ---- helpers -------------------------------------------------------------

func testCfg() *config.Config {
	return &config.Config{
		Port:            "3000",
		RateLimitRPS:    1000,
		RateLimitBurst:  1000,
		CacheTTLSeconds: 300,
	}
}

func do(t *testing.T, svc *testutil.MockService, method, path string, body []byte) *http.Response {
	t.Helper()
	app := api.NewApp(svc, io.Discard, testCfg())
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
	svc := testutil.NewMockService()
	body := []byte(`{"make":"Toyota","model":"Camry","year":2022,"color":"Blue","price":25000,"status":"available"}`)
	resp := do(t, svc, http.MethodPost, "/v1/cars", body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var car model.Car
	readJSON(t, resp, &car)
	assert.Equal(t, "Toyota", car.Make)
}

func TestCreateCar_ValidationError(t *testing.T) {
	svc := testutil.NewMockService()
	resp := do(t, svc, http.MethodPost, "/v1/cars", []byte(`{"make":"Toyota"}`))

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var er map[string]any
	readJSON(t, resp, &er)
	assert.Equal(t, "validation failed", er["error"])
}

func TestGetCar_Success(t *testing.T) {
	svc := testutil.NewMockService()
	svc.Cars["car-1"] = &model.Car{ID: "car-1", Make: "Honda"}
	resp := do(t, svc, http.MethodGet, "/v1/cars/car-1", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var car model.Car
	readJSON(t, resp, &car)
	assert.Equal(t, "Honda", car.Make)
}

func TestGetCar_NotFound(t *testing.T) {
	svc := testutil.NewMockService()
	resp := do(t, svc, http.MethodGet, "/v1/cars/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListCars(t *testing.T) {
	svc := testutil.NewMockService()
	svc.Cars["a"] = &model.Car{ID: "a", Make: "BMW"}
	svc.Cars["b"] = &model.Car{ID: "b", Make: "Audi"}
	resp := do(t, svc, http.MethodGet, "/v1/cars?page_size=10", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var result model.ListCarsResponse
	readJSON(t, resp, &result)
	assert.Len(t, result.Cars, 2)
}

func TestListCars_InvalidPageSize(t *testing.T) {
	svc := testutil.NewMockService()
	resp := do(t, svc, http.MethodGet, "/v1/cars?page_size=999", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateCar_Success(t *testing.T) {
	svc := testutil.NewMockService()
	svc.Cars["u1"] = &model.Car{ID: "u1", Make: "Ford"}
	resp := do(t, svc, http.MethodPut, "/v1/cars/u1", []byte(`{"make":"Tesla"}`))

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var car model.Car
	readJSON(t, resp, &car)
	assert.Equal(t, "Tesla", car.Make)
}

func TestUpdateCar_NotFound(t *testing.T) {
	svc := testutil.NewMockService()
	resp := do(t, svc, http.MethodPut, "/v1/cars/ghost", []byte(`{"make":"X"}`))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteCar_Success(t *testing.T) {
	svc := testutil.NewMockService()
	svc.Cars["d1"] = &model.Car{ID: "d1"}
	resp := do(t, svc, http.MethodDelete, "/v1/cars/d1", nil)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteCar_NotFound(t *testing.T) {
	svc := testutil.NewMockService()
	resp := do(t, svc, http.MethodDelete, "/v1/cars/phantom", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHealth(t *testing.T) {
	svc := testutil.NewMockService()
	resp := do(t, svc, http.MethodGet, "/health", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
