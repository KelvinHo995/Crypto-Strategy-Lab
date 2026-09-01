package sentiment

import (
	"context"
	"database/sql"
	"fmt"
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
			news_id, published_at, sentiment, score, model_name, model_version, analyzed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (news_id) DO UPDATE SET
			published_at = EXCLUDED.published_at,
			sentiment = EXCLUDED.sentiment,
			score = EXCLUDED.score,
			model_name = EXCLUDED.model_name,
			model_version = EXCLUDED.model_version,
			analyzed_at = EXCLUDED.analyzed_at
	`, observation.NewsID, observation.PublishedAt, observation.Sentiment,
		observation.Score, observation.ModelName, observation.ModelVersion, observation.AnalyzedAt)
	if err != nil {
		return fmt.Errorf("save sentiment observation: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListSince(ctx context.Context, earliestPublishedAt int64) ([]Observation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT news_id, published_at, sentiment, score, model_name, model_version, analyzed_at
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
		if err := rows.Scan(&o.NewsID, &o.PublishedAt, &o.Sentiment, &o.Score, &o.ModelName, &o.ModelVersion, &o.AnalyzedAt); err != nil {
			return nil, fmt.Errorf("scan sentiment observation: %w", err)
		}
		observations = append(observations, o)
	}
	return observations, rows.Err()
}
