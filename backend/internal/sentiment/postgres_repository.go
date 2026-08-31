package sentiment

import (
	"context"
	"database/sql"
	"errors"
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

func (r *PostgresRepository) LatestAtOrBefore(ctx context.Context, timestamp, earliestTimestamp int64) (Observation, error) {
	var observation Observation
	err := r.db.QueryRowContext(ctx, `
		SELECT news_id, published_at, sentiment, score, model_name, model_version, analyzed_at
		FROM sentiment_results
		WHERE published_at <= $1 AND published_at >= $2
		ORDER BY published_at DESC, analyzed_at DESC
		LIMIT 1
	`, timestamp, earliestTimestamp).Scan(
		&observation.NewsID,
		&observation.PublishedAt,
		&observation.Sentiment,
		&observation.Score,
		&observation.ModelName,
		&observation.ModelVersion,
		&observation.AnalyzedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Observation{}, ErrNotFound
	}
	if err != nil {
		return Observation{}, fmt.Errorf("lookup sentiment observation: %w", err)
	}
	return observation, nil
}
