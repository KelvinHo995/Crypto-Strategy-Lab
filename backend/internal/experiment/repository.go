package experiment

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("experiment not found")

type Repository interface {
	Save(ctx context.Context, r Result) error
	Get(ctx context.Context, id string) (Result, error)
	List(ctx context.Context) ([]Result, error)
}
