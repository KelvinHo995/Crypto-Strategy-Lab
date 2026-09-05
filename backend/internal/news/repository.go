package news

import "context"

// Repository stores normalized news independently from sentiment enrichment.
type Repository interface {
	Upsert(ctx context.Context, items []NewsItem) error
}
