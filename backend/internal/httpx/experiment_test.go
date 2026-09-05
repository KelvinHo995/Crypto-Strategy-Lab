package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type fakeRepo struct {
	mu      sync.Mutex
	results map[string]experiment.Result
	trades  map[string][]experiment.Trade
}

type fakeCandleRepo struct {
	symbol    string
	timeframe string
}

func (f *fakeCandleRepo) Upsert(context.Context, []market.Candle) error { return nil }
func (f *fakeCandleRepo) Range(_ context.Context, symbol, timeframe string, from, _ int64) ([]market.Candle, error) {
	f.symbol, f.timeframe = symbol, timeframe
	candles := make([]market.Candle, experiment.MinCandlesForBacktest)
	for i := range candles {
		candles[i] = market.Candle{Symbol: symbol, Timeframe: timeframe, OpenTime: from + int64(i), Open: 100, High: 101, Low: 99, Close: 100, Volume: 1, IsClosed: true}
	}
	return candles, nil
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{results: make(map[string]experiment.Result), trades: make(map[string][]experiment.Trade)}
}

func (f *fakeRepo) Save(_ context.Context, r experiment.Result) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.results[r.ID] = r
	return nil
}

func (f *fakeRepo) Get(_ context.Context, id string) (experiment.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.results[id]
	if !ok {
		return experiment.Result{}, experiment.ErrNotFound
	}
	return r, nil
}

func (f *fakeRepo) List(_ context.Context) ([]experiment.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]experiment.Result, 0, len(f.results))
	for _, r := range f.results {
		out = append(out, r)
	}
	return out, nil
}

func (f *fakeRepo) ListBySearch(_ context.Context, searchID string) ([]experiment.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []experiment.Result
	for _, r := range f.results {
		if r.SearchID == searchID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeRepo) SaveTrades(_ context.Context, experimentID string, trades []experiment.Trade) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.trades[experimentID] = trades
	return nil
}

func (f *fakeRepo) ListTrades(_ context.Context, experimentID string) ([]experiment.Trade, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.trades[experimentID], nil
}

func newTestRegistry() *strategy.Registry {
	reg := strategy.NewRegistry()
	reg.Register(strategy.NewMAStrategy(5, 10))
	reg.RegisterFactory("MA", strategy.MAFactory)
	return reg
}

func TestSearchStart_QueuesRealPipelineAndSavesProvenance(t *testing.T) {
	repo := newFakeRepo()
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), repo))
	defer srv.Close()

	body := `{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":10000000,"capital":1000,"instances":[{"type":"MA"}]}`
	resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", resp.StatusCode)
	}
	var out httpx.StartSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.SearchID == "" || out.Status != "STARTED" {
		t.Fatalf("got %+v", out)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		result, err := repo.Get(context.Background(), out.SearchID)
		if err != nil {
			t.Fatal(err)
		}
		if result.Status == "COMPLETED" {
			if result.StrategyVersions["MA"] != "v1" {
				t.Fatalf("strategyVersions = %#v, want MA=v1", result.StrategyVersions)
			}
			if len(result.SentimentModels) != 0 {
				t.Fatalf("non-sentiment experiment has sentiment models: %v", result.SentimentModels)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("search stayed in status %q", result.Status)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// End-to-end proof that /search/start accepts two instances of the same
// type with independently different params and custom weights, and
// persists exactly that in provenance — not collapsed to one shared bag or
// silently re-equalized weights, which is what the old strategies[]+params
// contract could never express in the first place.
func TestSearchStart_AcceptsMultipleInstancesOfSameTypeWithCustomWeights(t *testing.T) {
	repo := newFakeRepo()
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), repo))
	defer srv.Close()

	body := `{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":10000000,"capital":1000,"policy":"weighted",` +
		`"instances":[` +
		`{"type":"MA","params":{"maShortWindow":5,"maLongWindow":10},"weight":0.8},` +
		`{"type":"MA","params":{"maShortWindow":40,"maLongWindow":50},"weight":0.2}` +
		`]}`
	resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", resp.StatusCode)
	}
	var out httpx.StartSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}

	result, err := repo.Get(context.Background(), out.SearchID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Policy != "weighted" {
		t.Fatalf("Policy = %q, want weighted", result.Policy)
	}
	if len(result.Instances) != 2 {
		t.Fatalf("Instances = %d, want 2", len(result.Instances))
	}
	if result.Instances[0].Params["maShortWindow"] == result.Instances[1].Params["maShortWindow"] {
		t.Fatalf("both instances have the same maShortWindow=%v — params collapsed onto one shared bag", result.Instances[0].Params["maShortWindow"])
	}
	if result.Instances[0].Weight != 0.8 || result.Instances[1].Weight != 0.2 {
		t.Fatalf("weights = %v/%v, want 0.8/0.2 as submitted", result.Instances[0].Weight, result.Instances[1].Weight)
	}
}

func TestSearchStart_UnknownStrategy(t *testing.T) {
	repo := newFakeRepo()
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), repo))
	defer srv.Close()

	body := `{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":10000000,"capital":1000,"instances":[{"type":"NOPE"}]}`
	resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (unknown strategy name should be rejected)", resp.StatusCode)
	}
}

func TestSearchStart_UnsupportedMarket(t *testing.T) {
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), newFakeRepo()))
	defer srv.Close()

	body := `{"pair":"NOTREAL","timeframe":"5m","from":1,"to":10000000,"capital":1000,"instances":[{"type":"MA"}]}`
	resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for unsupported market", resp.StatusCode)
	}
}

func TestSearchStart_NormalizesMarketLookup(t *testing.T) {
	candles := &fakeCandleRepo{}
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Candles: candles})
	defer router.Close()
	srv := httptest.NewServer(router)
	defer srv.Close()

	body := `{"pair":" btcusdt ","timeframe":"5m","from":1,"to":1000,"capital":1000,"instances":[{"type":" MA "}]}`
	resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status=%d want 202", resp.StatusCode)
	}
	if candles.symbol != "BTCUSDT" || candles.timeframe != "5m" {
		t.Fatalf("lookup=%s/%s", candles.symbol, candles.timeframe)
	}
}

func TestSearchStart_RejectsUnknownFieldsAndTrailingJSON(t *testing.T) {
	for _, body := range []string{
		`{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"instances":[{"type":"MA"}],"surprise":true}`,
		`{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"instances":[{"type":"MA"}]} {}`,
	} {
		srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), newFakeRepo()))
		resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		srv.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 for %s", resp.StatusCode, body)
		}
	}
}

func TestSearchStart_RejectsFeeOrSlippageOutOfRange(t *testing.T) {
	for _, body := range []string{
		`{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"instances":[{"type":"MA"}],"fee":-0.1}`,
		`{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"instances":[{"type":"MA"}],"fee":2.5}`,
		`{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"instances":[{"type":"MA"}],"slippage":-1}`,
		`{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"instances":[{"type":"MA"}],"slippage":201}`,
	} {
		srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), newFakeRepo()))
		resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		srv.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 for %s", resp.StatusCode, body)
		}
	}
}

func TestSearchStart_AcceptsCustomFeeAndSlippage(t *testing.T) {
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), newFakeRepo()))
	defer srv.Close()

	body := `{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":1000,"capital":1000,"instances":[{"type":"MA"}],"fee":0.5,"slippage":25}`
	resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 (a fee/slippage within range must be accepted, not rejected as unknown)", resp.StatusCode)
	}
}

func TestExperimentsEmptyListIsJSONArray(t *testing.T) {
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), newFakeRepo()))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/experiments")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var results []experiment.Result
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatal(err)
	}
	if results == nil {
		t.Fatal("empty leaderboard encoded as null, want []")
	}
}

func TestGetExperiment_NotFound(t *testing.T) {
	repo := newFakeRepo()
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), repo))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/experiments/does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestGetExperiment_ExposesSentimentModelProvenance(t *testing.T) {
	repo := newFakeRepo()
	repo.results["sentiment-result"] = experiment.Result{
		ID:        "sentiment-result",
		Instances: []strategy.StrategyInstance{{Type: "Sentiment"}},
		StrategyVersions: map[string]string{
			"Sentiment": strategy.SentimentStrategyVersion,
		},
		SentimentModels: []strategy.SentimentModelIdentity{
			{Name: "crypto-lexicon", Version: "runtime-release"},
		},
		Status: "COMPLETED",
	}
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), repo))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/experiments/sentiment-result")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d, want 200", resp.StatusCode)
	}
	var result experiment.Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result.SentimentModels) != 1 || result.SentimentModels[0].Name != "crypto-lexicon" || result.SentimentModels[0].Version != "runtime-release" {
		t.Fatalf("sentiment models=%v", result.SentimentModels)
	}
}

func TestGetExperimentTrades_ReturnsRealTrades(t *testing.T) {
	repo := newFakeRepo()
	repo.trades["exp-1"] = []experiment.Trade{
		{Pair: "BTCUSDT", EntryTime: 1, Direction: experiment.Long, VolumeUSD: 1000, EntryPrice: 100, ExitPrice: 110, ExitTime: 2, Profit: 100},
	}
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), repo))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/experiments/exp-1/trades")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var trades []experiment.Trade
	if err := json.NewDecoder(resp.Body).Decode(&trades); err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 || trades[0].Profit != 100 {
		t.Fatalf("trades = %+v, want the one seeded trade", trades)
	}
}

func TestGetExperimentTrades_EmptyArrayNotNullForUnknownExperiment(t *testing.T) {
	repo := newFakeRepo()
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), repo))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/experiments/does-not-exist/trades")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (no trades yet is not an error)", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if strings.TrimSpace(string(body)) != "[]" {
		t.Fatalf("body = %q, want an empty JSON array, not null", body)
	}
}
