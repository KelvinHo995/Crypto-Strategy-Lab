package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type fakeRepo struct {
	mu      sync.Mutex
	results map[string]experiment.Result
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{results: make(map[string]experiment.Result)}
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

func newTestRegistry() *strategy.Registry {
	reg := strategy.NewRegistry()
	reg.Register(strategy.NewMAStrategy(5, 10))
	reg.RegisterFactory("MA", strategy.MAFactory)
	return reg
}

func TestSearchStart_RunsRealPipelineAndSaves(t *testing.T) {
	repo := newFakeRepo()
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), repo))
	defer srv.Close()

	body := `{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":10000000,"capital":1000,"strategies":["MA"]}`
	resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var out httpx.StartSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.SearchID == "" || out.Status != "COMPLETED" {
		t.Fatalf("got %+v", out)
	}

	getResp, err := http.Get(srv.URL + "/experiments/" + out.SearchID)
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET /experiments/{id} status = %d, want 200 (result should be saved and readable back)", getResp.StatusCode)
	}
}

func TestSearchStart_UnknownStrategy(t *testing.T) {
	repo := newFakeRepo()
	srv := httptest.NewServer(httpx.NewRouter(newTestRegistry(), repo))
	defer srv.Close()

	body := `{"pair":"BTCUSDT","timeframe":"5m","from":1,"to":10000000,"capital":1000,"strategies":["NOPE"]}`
	resp, err := http.Post(srv.URL+"/search/start", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (unknown strategy name should be rejected)", resp.StatusCode)
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
