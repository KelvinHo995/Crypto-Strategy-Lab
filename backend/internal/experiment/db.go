package experiment

import (
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// OpenDB opens a Postgres connection pool. dsn is the connection string,
// e.g. from the DATABASE_URL env var — see migrations/0001_init.sql and
// ADR-0012 for the schema/setup this expects.
//
// Uses simple query protocol, not pgx's default server-side prepared
// statement caching: Supabase's pooled connection string (port 6543,
// transaction mode) can route different requests to different backend
// connections, so a cached prepared statement name from one request can
// already exist on the backend a later request lands on ("prepared
// statement already exists", SQLSTATE 42P05) — this surfaces under
// concurrent load, which is exactly what the worker pool produces.
//
// Pool size is capped well under Supabase's pooler budget (Project
// Settings > Database > Connection Pooling: 15 connections, shared across
// every pooler and every team member's local backend instance) so one
// instance can't starve the others.
func OpenDB(dsn string) (*sql.DB, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	db := stdlib.OpenDB(*config)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
