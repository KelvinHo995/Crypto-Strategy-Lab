package sentiment_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
)

func TestTimeLookupMapsModelConfidenceToDirectionalScore(t *testing.T) {
	tests := []struct {
		label      string
		confidence float64
		want       float64
	}{
		{label: "POSITIVE", confidence: 0.9, want: 0.9},
		{label: "NEGATIVE", confidence: 0.9, want: 0.1},
		{label: "NEUTRAL", confidence: 0.9, want: 0.5},
	}

	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			repo := &fakeRepository{observation: sentiment.Observation{
				Sentiment: test.label,
				Score:     test.confidence,
			}}
			lookup := sentiment.NewTimeLookup(repo, 24*time.Hour)
			got, err := lookup.FetchSentiment(context.Background(), 200_000_000)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got-test.want) > 1e-9 {
				t.Fatalf("score = %v, want %v", got, test.want)
			}
		})
	}
}

func TestTimeLookupBoundsRepositoryQuery(t *testing.T) {
	repo := &fakeRepository{observation: sentiment.Observation{Sentiment: "NEUTRAL", Score: 0.5}}
	lookup := sentiment.NewTimeLookup(repo, 2*time.Hour)
	timestamp := int64(10_000_000)
	if _, err := lookup.FetchSentiment(context.Background(), timestamp); err != nil {
		t.Fatal(err)
	}
	if repo.latest != timestamp {
		t.Fatalf("latest timestamp = %d, want %d", repo.latest, timestamp)
	}
	wantEarliest := timestamp - (2 * time.Hour).Milliseconds()
	if repo.earliest != wantEarliest {
		t.Fatalf("earliest timestamp = %d, want %d", repo.earliest, wantEarliest)
	}
}

func TestTimeLookupPreservesMissingObservationError(t *testing.T) {
	lookup := sentiment.NewTimeLookup(&fakeRepository{err: sentiment.ErrNotFound}, time.Hour)
	if _, err := lookup.FetchSentiment(context.Background(), 10_000_000); !errors.Is(err, sentiment.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestTimeLookupCapsMaximumAge(t *testing.T) {
	repo := &fakeRepository{observation: sentiment.Observation{Sentiment: "NEUTRAL", Score: 0.5}}
	lookup := sentiment.NewTimeLookup(repo, 7*24*time.Hour)
	timestamp := int64(200_000_000)
	if _, err := lookup.FetchSentiment(context.Background(), timestamp); err != nil {
		t.Fatal(err)
	}
	wantEarliest := timestamp - sentiment.DefaultMaxAge.Milliseconds()
	if repo.earliest != wantEarliest {
		t.Fatalf("earliest timestamp = %d, want capped value %d", repo.earliest, wantEarliest)
	}
}
