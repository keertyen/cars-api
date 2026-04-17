package model

import "time"

// CarStatus is the availability state of a car.
type CarStatus string

const (
	StatusAvailable CarStatus = "available"
	StatusSold      CarStatus = "sold"
	StatusPending   CarStatus = "pending"
)

// Car is the canonical domain object.
// The ID field carries the Firestore document ID and is excluded from the
// stored document body (firestore:"-").
type Car struct {
	ID        string    `json:"id"                  firestore:"-"`
	Make      string    `json:"make"                firestore:"make"`
	Model     string    `json:"model"               firestore:"model"`
	Year      int       `json:"year"                firestore:"year"`
	Color     string    `json:"color"               firestore:"color"`
	Price     float64   `json:"price"               firestore:"price"`
	Mileage   int       `json:"mileage"             firestore:"mileage"`
	VIN       string    `json:"vin,omitempty"       firestore:"vin"`
	Status    CarStatus `json:"status"              firestore:"status"`
	CreatedAt time.Time `json:"created_at"          firestore:"created_at"`
	UpdatedAt time.Time `json:"updated_at"          firestore:"updated_at"`
}

// CreateCarRequest is validated on ingestion before reaching the service layer.
type CreateCarRequest struct {
	Make    string    `json:"make"          validate:"required,min=1,max=100"`
	Model   string    `json:"model"         validate:"required,min=1,max=100"`
	Year    int       `json:"year"          validate:"required,min=1886,max=2030"`
	Color   string    `json:"color"         validate:"required,min=1,max=50"`
	Price   float64   `json:"price"         validate:"required,gt=0"`
	Mileage int       `json:"mileage"       validate:"min=0"`
	VIN     string    `json:"vin,omitempty" validate:"omitempty,len=17,alphanum"`
	Status  CarStatus `json:"status"        validate:"required,oneof=available sold pending"`
}

// UpdateCarRequest uses pointer fields so the service can distinguish
// "not provided" from "explicitly set to zero value".
type UpdateCarRequest struct {
	Make    *string    `json:"make,omitempty"    validate:"omitempty,min=1,max=100"`
	Model   *string    `json:"model,omitempty"   validate:"omitempty,min=1,max=100"`
	Year    *int       `json:"year,omitempty"    validate:"omitempty,min=1886,max=2030"`
	Color   *string    `json:"color,omitempty"   validate:"omitempty,min=1,max=50"`
	Price   *float64   `json:"price,omitempty"   validate:"omitempty,gt=0"`
	Mileage *int       `json:"mileage,omitempty" validate:"omitempty,min=0"`
	VIN     *string    `json:"vin,omitempty"     validate:"omitempty,len=17,alphanum"`
	Status  *CarStatus `json:"status,omitempty"  validate:"omitempty,oneof=available sold pending"`
}

// ListCarsQuery carries pagination parameters for the store layer.
type ListCarsQuery struct {
	PageSize  int    `validate:"min=1,max=100"`
	PageToken string // last document ID — used as cursor for Firestore StartAfter
}

// ListCarsResponse is the envelope returned by the list endpoint.
type ListCarsResponse struct {
	Cars          []*Car `json:"cars"`
	NextPageToken string `json:"next_page_token,omitempty"`
}
