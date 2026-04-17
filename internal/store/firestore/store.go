package firestore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/yenug1k/cars-api/internal/model"
	"github.com/yenug1k/cars-api/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	collection  = "cars"
	opTimeout   = 5 * time.Second
	listTimeout = 10 * time.Second
)

type Store struct {
	client *firestore.Client
}

func New(ctx context.Context, projectID string) (*Store, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("firestore.NewClient: %w", err)
	}
	return &Store{client: client}, nil
}

func (s *Store) Close() error {
	return s.client.Close()
}

func (s *Store) Create(ctx context.Context, car *model.Car) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	_, err := s.client.Collection(collection).Doc(car.ID).Set(ctx, car)
	if err != nil {
		return fmt.Errorf("firestore Create: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (*model.Car, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	doc, err := s.client.Collection(collection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("firestore Get: %w", err)
	}

	var car model.Car
	if err := doc.DataTo(&car); err != nil {
		return nil, fmt.Errorf("firestore DataTo: %w", err)
	}
	car.ID = doc.Ref.ID
	return &car, nil
}

func (s *Store) List(ctx context.Context, q model.ListCarsQuery) ([]*model.Car, error) {
	ctx, cancel := context.WithTimeout(ctx, listTimeout)
	defer cancel()

	query := s.client.Collection(collection).
		OrderBy("created_at", firestore.Desc).
		Limit(q.PageSize)

	// StartAfter avoids Firestore's O(offset) read cost.
	if q.PageToken != "" {
		anchorDoc, err := s.client.Collection(collection).Doc(q.PageToken).Get(ctx)
		if err == nil {
			query = query.StartAfter(anchorDoc)
		}
		// If the anchor doc is gone we silently fall back to the first page.
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore List: %w", err)
	}

	cars := make([]*model.Car, 0, len(docs))
	for _, doc := range docs {
		var car model.Car
		if err := doc.DataTo(&car); err != nil {
			return nil, fmt.Errorf("firestore DataTo %s: %w", doc.Ref.ID, err)
		}
		car.ID = doc.Ref.ID
		cars = append(cars, &car)
	}
	return cars, nil
}

func (s *Store) Update(ctx context.Context, car *model.Car) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	_, err := s.client.Collection(collection).Doc(car.ID).Set(ctx, car)
	if err != nil {
		return fmt.Errorf("firestore Update: %w", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	_, err := s.client.Collection(collection).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("firestore Delete: %w", err)
	}
	return nil
}
