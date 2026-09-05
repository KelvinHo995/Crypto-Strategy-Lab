package news

import "context"

type NewsItem struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Text         string   `json:"text"`
	Source       string   `json:"source"`
	URL          string   `json:"url,omitempty"`
	PublishedAt  int64    `json:"publishedAt"`
	RelatedCoins []string `json:"relatedCoins"`
}

// NewsProvider supplies source-neutral news items for downstream processing.
type NewsProvider interface {
	Fetch(ctx context.Context, sinceTimestamp int64) ([]NewsItem, error)
}
