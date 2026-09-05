package news

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Upsert(ctx context.Context, items []NewsItem) error {
	if len(items) == 0 {
		return nil
	}
	query, args, err := newsUpsertStatement(items, time.Now().UnixMilli())
	if err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("upsert news items: %w", err)
	}
	return nil
}

func newsUpsertStatement(items []NewsItem, createdAt int64) (string, []any, error) {
	var query strings.Builder
	query.WriteString(`INSERT INTO news_items (id,title,content,source,url,published_at,related_coins,created_at) VALUES `)
	args := make([]any, 0, len(items)*8)
	for i, item := range items {
		if i > 0 {
			query.WriteByte(',')
		}
		query.WriteByte('(')
		for field := 0; field < 8; field++ {
			if field > 0 {
				query.WriteByte(',')
			}
			query.WriteByte('$')
			query.WriteString(strconv.Itoa(i*8 + field + 1))
		}
		query.WriteByte(')')

		relatedCoins := item.RelatedCoins
		if relatedCoins == nil {
			relatedCoins = []string{}
		}
		encodedCoins, err := json.Marshal(relatedCoins)
		if err != nil {
			return "", nil, fmt.Errorf("encode related coins for %q: %w", item.ID, err)
		}
		args = append(args, item.ID, item.Title, item.Text, item.Source, item.URL,
			item.PublishedAt, string(encodedCoins), createdAt)
	}
	query.WriteString(` ON CONFLICT (id) DO UPDATE SET
		title=EXCLUDED.title,
		content=EXCLUDED.content,
		source=EXCLUDED.source,
		url=EXCLUDED.url,
		published_at=EXCLUDED.published_at,
		related_coins=EXCLUDED.related_coins`)
	return query.String(), args, nil
}

var _ Repository = (*PostgresRepository)(nil)
