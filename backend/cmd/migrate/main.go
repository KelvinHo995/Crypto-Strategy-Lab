package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	if os.Getenv("DATABASE_URL") == "" {
		_ = godotenv.Load("backend/.env")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	directory := "migrations"
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		directory = filepath.Join("backend", "migrations")
	}
	files, err := migrationFiles(directory)
	if err != nil {
		log.Fatal(err)
	}
	db, err := experiment.OpenDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for _, path := range files {
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			log.Fatal(readErr)
		}
		if _, execErr := db.ExecContext(ctx, string(contents)); execErr != nil {
			log.Fatalf("apply %s: %v", filepath.Base(path), execErr)
		}
		log.Printf("applied %s", filepath.Base(path))
	}
}

func migrationFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !isMigrationName(name) {
			continue
		}
		paths = append(paths, filepath.Join(directory, name))
	}
	if len(paths) == 0 {
		return nil, errors.New("no SQL migrations found")
	}
	return paths, nil
}

func isMigrationName(name string) bool {
	if len(name) < len("0001_.sql") || !strings.HasSuffix(name, ".sql") || name[4] != '_' {
		return false
	}
	for _, character := range name[:4] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
