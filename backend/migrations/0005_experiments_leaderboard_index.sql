CREATE INDEX IF NOT EXISTS experiments_leaderboard_score_idx
ON experiments (
    (CASE WHEN status = 'COMPLETED' THEN 0 ELSE 1 END),
    ((0.50 * COALESCE(return_pct, 0) + 0.30 * COALESCE(win_rate, 0) - 0.20 * COALESCE(mdd, 0))) DESC,
    created_at DESC
);
