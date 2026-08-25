package experiment

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenDB opens a Postgres connection pool. dsn is the connection string,
// e.g. from the DATABASE_URL env var — see migrations/0001_init.sql and
// ADR-0012 for the schema/setup this expects.
func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
