//go:build integration

package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenug1k/cars-api/config"
	"github.com/yenug1k/cars-api/internal/api"
	"github.com/yenug1k/cars-api/internal/model"
	"github.com/yenug1k/cars-api/internal/testutil"
)

func newApp(svc *testutil.MockService) *fiber.App {
	cfg := &config.Config{RateLimitBurst: 1000}
	return api.NewApp(svc, io.Discard, cfg)
}

func do(t *testing.T, app *fiber.App, method, path string, body []byte) *http.Response {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func TestHealth(t *testing.T) {
	app := newApp(testutil.NewMockService())
	resp := do(t, app, http.MethodGet, "/health", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateAndGetCar(t *testing.T) {
	app := newApp(testutil.NewMockService())

	body, _ := json.Marshal(map[string]any{
		"make": "Toyota", "model": "Camry", "year": 2022,
		"color": "white", "price": 25000.0, "status": "available",
	})

	resp := do(t, app, http.MethodPost, "/v1/cars/", body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created model.Car
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, "Toyota", created.Make)

	resp2 := do(t, app, http.MethodGet, "/v1/cars/"+created.ID, nil)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var fetched model.Car
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&fetched))
	assert.Equal(t, created.ID, fetched.ID)
}

func TestGetCar_NotFound(t *testing.T) {
	app := newApp(testutil.NewMockService())
	resp := do(t, app, http.MethodGet, "/v1/cars/does-not-exist", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListCars(t *testing.T) {
	app := newApp(testutil.NewMockService())

	for _, make_ := range []string{"Honda", "Ford"} {
		body, _ := json.Marshal(map[string]any{
			"make": make_, "model": "X", "year": 2020,
			"color": "black", "price": 20000.0, "status": "available",
		})
		resp := do(t, app, http.MethodPost, "/v1/cars/", body)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	resp := do(t, app, http.MethodGet, "/v1/cars/?page_size=10", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp model.ListCarsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResp))
	assert.Len(t, listResp.Cars, 2)
}

func TestUpdateCar(t *testing.T) {
	app := newApp(testutil.NewMockService())

	body, _ := json.Marshal(map[string]any{
		"make": "BMW", "model": "3 Series", "year": 2021,
		"color": "blue", "price": 45000.0, "status": "available",
	})
	resp := do(t, app, http.MethodPost, "/v1/cars/", body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created model.Car
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	update, _ := json.Marshal(map[string]any{"make": "Mercedes"})
	resp2 := do(t, app, http.MethodPut, "/v1/cars/"+created.ID, update)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var updated model.Car
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&updated))
	assert.Equal(t, "Mercedes", updated.Make)
}

func TestDeleteCar(t *testing.T) {
	app := newApp(testutil.NewMockService())

	body, _ := json.Marshal(map[string]any{
		"make": "Audi", "model": "A4", "year": 2023,
		"color": "silver", "price": 50000.0, "status": "available",
	})
	resp := do(t, app, http.MethodPost, "/v1/cars/", body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created model.Car
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	resp2 := do(t, app, http.MethodDelete, "/v1/cars/"+created.ID, nil)
	assert.Equal(t, http.StatusNoContent, resp2.StatusCode)

	resp3 := do(t, app, http.MethodGet, "/v1/cars/"+created.ID, nil)
	assert.Equal(t, http.StatusNotFound, resp3.StatusCode)
}

func TestCreateCar_ValidationError(t *testing.T) {
	app := newApp(testutil.NewMockService())

	body, _ := json.Marshal(map[string]any{"make": "Toyota"})
	resp := do(t, app, http.MethodPost, "/v1/cars/", body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
