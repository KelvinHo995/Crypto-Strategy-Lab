package experiment

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type capturingResultExecer struct {
	query string
	args  []any
}

func (e *capturingResultExecer) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	e.query = query
	e.args = append([]any(nil), args...)
	return staticSQLResult{}, nil
}

type staticSQLResult struct{}

func (staticSQLResult) LastInsertId() (int64, error) { return 0, nil }
func (staticSQLResult) RowsAffected() (int64, error) { return 1, nil }

type capturedResultRow struct {
	args []any
}

func (r capturedResultRow) Scan(dest ...any) error {
	if len(dest) != 21 || len(r.args) != 21 {
		return fmt.Errorf("destinations=%d args=%d, want 21", len(dest), len(r.args))
	}
	*(dest[0].(*string)) = r.args[0].(string)
	*(dest[1].(*string)) = r.args[1].(string)
	*(dest[2].(*int)) = r.args[2].(int)
	*(dest[3].(*string)) = r.args[3].(string)
	*(dest[4].(*string)) = r.args[4].(string)
	*(dest[5].(*string)) = r.args[5].(string)
	*(dest[6].(*[]byte)) = []byte(r.args[6].(string))
	*(dest[7].(*string)) = r.args[7].(string)
	*(dest[8].(*[]byte)) = []byte(r.args[8].(string))
	*(dest[9].(*[]byte)) = []byte(r.args[9].(string))
	*(dest[10].(*string)) = r.args[10].(string)
	*(dest[11].(*float64)) = r.args[11].(float64)
	*(dest[12].(*float64)) = r.args[12].(float64)
	*(dest[13].(*int)) = r.args[13].(int)
	*(dest[14].(*float64)) = r.args[14].(float64)
	*(dest[15].(*int)) = r.args[15].(int)
	*(dest[16].(*int)) = r.args[16].(int)
	*(dest[17].(*float64)) = r.args[17].(float64)
	*(dest[18].(*string)) = r.args[18].(string)
	*(dest[19].(*int64)) = r.args[19].(int64)
	*(dest[20].(*int64)) = r.args[20].(int64)
	return nil
}

func TestNormalizeLegacyJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain JSON", input: `["MA","RSI"]`, want: `["MA","RSI"]`},
		{name: "legacy bytea text", input: `\x5b224d41225d`, want: `["MA"]`},
		{name: "malformed bytea text", input: `\xnot-hex`, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeLegacyJSON([]byte(test.input))
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && !bytes.Equal(got, []byte(test.want)) {
				t.Fatalf("value = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSentimentModelProvenanceSurvivesRepositoryRoundTrip(t *testing.T) {
	execer := &capturingResultExecer{}
	want := Result{
		ID: "experiment-sentiment", SearchID: "search-1", SearchTotal: 1,
		CandidateID: "candidate-1",
		Pair:        "BTCUSDT",
		Timeframe:   "1h",
		Instances:   []strategy.StrategyInstance{{Type: "Sentiment"}},
		Policy:      "majority",
		StrategyVersions: map[string]string{
			"Sentiment": strategy.SentimentStrategyVersion,
		},
		SentimentModels: []strategy.SentimentModelIdentity{
			{Name: "crypto-lexicon", Version: "runtime-release"},
			{Name: "crypto-lexicon", Version: "runtime-release-2"},
		},
		DatasetPeriod: "1-2", Return: 1.2, MDD: 0.3, TradeCount: 2,
		WinRate: 0.5, Wins: 1, Losses: 1, TotalProfit: 12, Status: "COMPLETED", CreatedAt: 100,
	}
	if err := saveResult(context.Background(), execer, want); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(execer.query, "sentiment_models") {
		t.Fatalf("save query does not persist sentiment_models: %s", execer.query)
	}
	got, err := scanResult(capturedResultRow{args: execer.args})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.SentimentModels) != 2 || got.SentimentModels[0] != want.SentimentModels[0] || got.SentimentModels[1] != want.SentimentModels[1] {
		t.Fatalf("sentiment models=%v, want %v", got.SentimentModels, want.SentimentModels)
	}
	if got.StrategyVersions["Sentiment"] != strategy.SentimentStrategyVersion || got.Instances[0].Type != "Sentiment" {
		t.Fatalf("strategy provenance did not survive: %+v", got)
	}
	if got.Pair != want.Pair || got.Timeframe != want.Timeframe || got.DatasetPeriod != want.DatasetPeriod {
		t.Fatalf("market provenance=%s/%s/%s, want %s/%s/%s", got.Pair, got.Timeframe, got.DatasetPeriod, want.Pair, want.Timeframe, want.DatasetPeriod)
	}
}

func TestMarketContextRoundTripKeepsTimeframesDistinct(t *testing.T) {
	for _, timeframe := range []string{"1h", "15m"} {
		t.Run(timeframe, func(t *testing.T) {
			execer := &capturingResultExecer{}
			want := Result{
				ID: "experiment-" + timeframe, CandidateID: "candidate", Pair: "BTCUSDT", Timeframe: timeframe,
				Instances: []strategy.StrategyInstance{}, StrategyVersions: map[string]string{}, DatasetPeriod: "1-2",
				Status: "COMPLETED", CreatedAt: 100,
			}
			if err := saveResult(context.Background(), execer, want); err != nil {
				t.Fatal(err)
			}
			got, err := scanResult(capturedResultRow{args: execer.args})
			if err != nil {
				t.Fatal(err)
			}
			if got.Pair != "BTCUSDT" || got.Timeframe != timeframe {
				t.Fatalf("market context=%s/%s, want BTCUSDT/%s", got.Pair, got.Timeframe, timeframe)
			}
		})
	}
}

func TestMissingLegacyMarketContextRemainsUnknown(t *testing.T) {
	execer := &capturingResultExecer{}
	want := Result{
		ID: "legacy-experiment", CandidateID: "candidate", Instances: []strategy.StrategyInstance{},
		StrategyVersions: map[string]string{}, DatasetPeriod: "1-2", Status: "COMPLETED", CreatedAt: 100,
	}
	if err := saveResult(context.Background(), execer, want); err != nil {
		t.Fatal(err)
	}
	got, err := scanResult(capturedResultRow{args: execer.args})
	if err != nil {
		t.Fatal(err)
	}
	if got.Pair != "" || got.Timeframe != "" {
		t.Fatalf("legacy market context=%q/%q, want unknown empty values", got.Pair, got.Timeframe)
	}
}
