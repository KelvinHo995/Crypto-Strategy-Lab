package main

import (
	"testing"
	"time"
)

func TestParseWorkers(t *testing.T) {
	workers, err := parseWorkers("1, 3,6")
	if err != nil || len(workers) != 3 || workers[0] != 1 || workers[2] != 6 {
		t.Fatalf("workers=%v err=%v", workers, err)
	}
	if _, err := parseWorkers("0"); err == nil {
		t.Fatal("zero workers accepted")
	}
}

func TestMedianDuration(t *testing.T) {
	if got := medianDuration([]time.Duration{30 * time.Millisecond, 10 * time.Millisecond, 20 * time.Millisecond}); got != 20*time.Millisecond {
		t.Fatalf("odd median=%s", got)
	}
	if got := medianDuration([]time.Duration{40 * time.Millisecond, 10 * time.Millisecond}); got != 25*time.Millisecond {
		t.Fatalf("even median=%s", got)
	}
}
