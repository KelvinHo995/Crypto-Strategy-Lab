package main

import "testing"

func TestMigrationName(t *testing.T) {
	for _, name := range []string{"0001_init.sql", "0006_experiment_jobs.sql", "0007_experiment_search_metadata.sql"} {
		if !isMigrationName(name) {
			t.Fatalf("%s should be accepted", name)
		}
	}
	for _, name := range []string{"notes.sql", "001_bad.sql", "0001_init.txt"} {
		if isMigrationName(name) {
			t.Fatalf("%s should be rejected", name)
		}
	}
}
