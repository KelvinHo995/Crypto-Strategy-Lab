package httpx

import (
	"context"
	"net/http"
	"sync"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/auth"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

// NewRouter wires public routes directly and mounts everything else behind
// requireAuth on a single catch-all — one gate, not a per-route list.
type Router struct {
	handler   http.Handler
	cancel    context.CancelFunc
	pool      *experiment.WorkerPool
	closeOnce sync.Once
}
type Dependencies struct {
	Auth      *auth.Service
	Candles   market.CandleRepository
	Live      market.LiveProvider
	Sentiment SentimentService
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) { r.handler.ServeHTTP(w, req) }

func (r *Router) Close() {
	r.closeOnce.Do(func() {
		r.cancel()
		r.pool.Wait()
	})
}

func NewRouter(registry *strategy.Registry, repo experiment.Repository) *Router {
	return NewRouterWithContext(context.Background(), registry, repo)
}

func NewRouterWithContext(parent context.Context, registry *strategy.Registry, repo experiment.Repository, dependencies ...Dependencies) *Router {
	var deps Dependencies
	if len(dependencies) > 0 {
		deps = dependencies[0]
	}
	authService := deps.Auth
	ctx, cancel := context.WithCancel(parent)
	queue := experiment.NewInMemoryQueue(128)
	pool := experiment.NewWorkerPool(queue, registry, repo, 3)
	hub := NewHub(repo)
	pool.SetObserver(hub.JobUpdated)
	pool.Start(ctx)
	if deps.Live != nil {
		for _, frame := range []string{"5m", "15m", "1h", "4h"} {
			candles := deps.Live.StreamLiveCandles(ctx, "BTCUSDT", frame)
			go func() {
				for c := range candles {
					hub.Broadcast(Event{Type: "CANDLE_UPDATE", Payload: c})
				}
			}()
		}
	}
	public := http.NewServeMux()
	public.HandleFunc("GET /health", health)
	if authService != nil {
		public.HandleFunc("POST /auth/register", register(authService))
		public.HandleFunc("POST /auth/login", login(authService))
	}

	protected := http.NewServeMux()
	protected.HandleFunc("POST /auth/logout", logout)
	strategyHandler := strategy.NewHandler(registry)

	protected.HandleFunc("POST /search/start", startSearch(registry, repo, queue, deps.Candles))
	protected.HandleFunc("GET /experiments", listExperiments(repo))
	protected.HandleFunc("GET /experiments/{id}", getExperiment(repo))
	protected.HandleFunc("GET /strategies", strategyHandler.ListStrategies)
	protected.HandleFunc("GET /candles", listCandles(deps.Candles))
	protected.HandleFunc("POST /sentiment/analyze", analyzeSentiment(deps.Sentiment))
	protected.HandleFunc("GET /ws", serveWebSocket(hub))

	public.Handle("/", requireAuth(authService, protected))
	return &Router{handler: public, cancel: cancel, pool: pool}
}
