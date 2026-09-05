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

func TestTimeLookupReturnsPersistedModelIdentity(t *testing.T) {
	repo := &fakeRepository{observation: sentiment.Observation{
		Sentiment:    "POSITIVE",
		Score:        0.91,
		ModelName:    "crypto-lexicon",
		ModelVersion: "runtime-release",
	}}
	lookup := sentiment.NewTimeLookup(repo, time.Hour)

	score, name, version, err := lookup.FetchSentimentWithModel(context.Background(), 10_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if score != 0.91 || name != "crypto-lexicon" || version != "runtime-release" {
		t.Fatalf("lookup = (%v, %q, %q), want persisted identity", score, name, version)
	}
}

func TestTimeLookupBoundsRepositoryQuery(t *testing.T) {
	repo := &fakeRepository{observation: sentiment.Observation{Sentiment: "NEUTRAL", Score: 0.5}}
	lookup := sentiment.NewTimeLookup(repo, 2*time.Hour)
	timestamp := int64(10_000_000)
	if _, err := lookup.FetchSentiment(context.Background(), timestamp); err != nil {
		t.Fatal(err)
	}
	wantEarliest := timestamp - (2 * time.Hour).Milliseconds()
	if repo.earliestSince != wantEarliest {
		t.Fatalf("earliestSince = %d, want %d", repo.earliestSince, wantEarliest)
	}
}

// Within one backtest, candle timestamps are asked about in ascending order
// — the whole point of the range-based cache is that the second and later
// lookups reuse the first fetch instead of re-querying per candle.
func TestTimeLookupCachesAcrossAscendingLookups(t *testing.T) {
	repo := &fakeRepository{observation: sentiment.Observation{Sentiment: "NEUTRAL", Score: 0.5, PublishedAt: 9_999_999}}
	lookup := sentiment.NewTimeLookup(repo, 24*time.Hour)

	if _, err := lookup.FetchSentiment(context.Background(), 10_000_000); err != nil {
		t.Fatal(err)
	}
	repo.earliestSince = -1 // sentinel: a real ListSince call would overwrite this
	if _, err := lookup.FetchSentiment(context.Background(), 10_000_001); err != nil {
		t.Fatal(err)
	}
	if repo.earliestSince != -1 {
		t.Fatal("ListSince called again for a timestamp within the cached range — cache was not reused")
	}
}

// Invalidate is how a writer tells the cache "don't trust what you have" —
// without it, new observations wouldn't be visible until the TTL expires.
func TestTimeLookupInvalidateForcesRefetch(t *testing.T) {
	repo := &fakeRepository{observation: sentiment.Observation{Sentiment: "NEUTRAL", Score: 0.5}}
	lookup := sentiment.NewTimeLookup(repo, 24*time.Hour)

	if _, err := lookup.FetchSentiment(context.Background(), 10_000_000); err != nil {
		t.Fatal(err)
	}
	repo.earliestSince = -1
	lookup.Invalidate()
	if _, err := lookup.FetchSentiment(context.Background(), 10_000_001); err != nil {
		t.Fatal(err)
	}
	if repo.earliestSince == -1 {
		t.Fatal("ListSince was not called again after Invalidate")
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
	if repo.earliestSince != wantEarliest {
		t.Fatalf("earliestSince = %d, want capped value %d", repo.earliestSince, wantEarliest)
	}
}
