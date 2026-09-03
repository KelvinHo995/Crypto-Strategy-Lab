package sentiment

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, observation Observation) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sentiment_results (
			news_id, title, source, url, published_at, sentiment, score, model_name, model_version, analyzed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (news_id) DO UPDATE SET
			title = EXCLUDED.title,
			source = EXCLUDED.source,
			url = EXCLUDED.url,
			published_at = EXCLUDED.published_at,
			sentiment = EXCLUDED.sentiment,
			score = EXCLUDED.score,
			model_name = EXCLUDED.model_name,
			model_version = EXCLUDED.model_version,
			analyzed_at = EXCLUDED.analyzed_at
	`, observation.NewsID, observation.Title, observation.Source, observation.URL, observation.PublishedAt,
		observation.Sentiment, observation.Score, observation.ModelName, observation.ModelVersion, observation.AnalyzedAt)
	if err != nil {
		return fmt.Errorf("save sentiment observation: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ExistingNewsIDs(ctx context.Context, newsIDs []string) (map[string]struct{}, error) {
	existing := make(map[string]struct{})
	if len(newsIDs) == 0 {
		return existing, nil
	}

	placeholders := make([]string, len(newsIDs))
	args := make([]any, len(newsIDs))
	for i, newsID := range newsIDs {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = newsID
	}
	query := `SELECT news_id FROM sentiment_results WHERE news_id IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("check existing sentiment observations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var newsID string
		if err := rows.Scan(&newsID); err != nil {
			return nil, fmt.Errorf("scan existing sentiment observation: %w", err)
		}
		existing[newsID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read existing sentiment observations: %w", err)
	}
	return existing, nil
}

func (r *PostgresRepository) ListSince(ctx context.Context, earliestPublishedAt int64) ([]Observation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT news_id, title, source, url, published_at, sentiment, score, model_name, model_version, analyzed_at
		FROM sentiment_results
		WHERE published_at >= $1
		ORDER BY published_at ASC
	`, earliestPublishedAt)
	if err != nil {
		return nil, fmt.Errorf("list sentiment observations: %w", err)
	}
	defer rows.Close()

	var observations []Observation
	for rows.Next() {
		var o Observation
		if err := rows.Scan(&o.NewsID, &o.Title, &o.Source, &o.URL, &o.PublishedAt, &o.Sentiment, &o.Score, &o.ModelName, &o.ModelVersion, &o.AnalyzedAt); err != nil {
			return nil, fmt.Errorf("scan sentiment observation: %w", err)
		}
		observations = append(observations, o)
	}
	return observations, rows.Err()
}
