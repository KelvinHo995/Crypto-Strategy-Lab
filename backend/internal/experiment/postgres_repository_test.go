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
	if len(dest) != 19 || len(r.args) != 19 {
		return fmt.Errorf("destinations=%d args=%d, want 19", len(dest), len(r.args))
	}
	*(dest[0].(*string)) = r.args[0].(string)
	*(dest[1].(*string)) = r.args[1].(string)
	*(dest[2].(*int)) = r.args[2].(int)
	*(dest[3].(*string)) = r.args[3].(string)
	*(dest[4].(*[]byte)) = []byte(r.args[4].(string))
	*(dest[5].(*string)) = r.args[5].(string)
	*(dest[6].(*[]byte)) = []byte(r.args[6].(string))
	*(dest[7].(*[]byte)) = []byte(r.args[7].(string))
	*(dest[8].(*string)) = r.args[8].(string)
	*(dest[9].(*float64)) = r.args[9].(float64)
	*(dest[10].(*float64)) = r.args[10].(float64)
	*(dest[11].(*int)) = r.args[11].(int)
	*(dest[12].(*float64)) = r.args[12].(float64)
	*(dest[13].(*int)) = r.args[13].(int)
	*(dest[14].(*int)) = r.args[14].(int)
	*(dest[15].(*float64)) = r.args[15].(float64)
	*(dest[16].(*string)) = r.args[16].(string)
	*(dest[17].(*int64)) = r.args[17].(int64)
	*(dest[18].(*int64)) = r.args[18].(int64)
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
}
