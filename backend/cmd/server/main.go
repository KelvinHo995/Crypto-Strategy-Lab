package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/auth"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
	"github.com/joho/godotenv"
)

func main() {
	loadLocalEnv()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	registry := strategy.NewRegistry()
	mustRegisterPlugin(registry, strategy.NewMAPlugin(strategy.NewMAStrategy(20, 50)))
	mustRegisterPlugin(registry, strategy.NewRSIPlugin(strategy.NewRSIStrategy(14, 70.0, 30.0)))
	mustRegisterPlugin(registry, strategy.NewBollingerPlugin(strategy.NewBollingerStrategy(20, 2.0)))
	mustRegisterPlugin(registry, strategy.NewSRPlugin(strategy.NewSRStrategy(20, 0.005)))
	mustRegisterPlugin(registry, strategy.NewSMCPlugin(strategy.NewSMCStrategy(10)))
	mustRegisterPlugin(registry, strategy.NewMACDPlugin(strategy.NewMACDStrategy(12, 26, 9)))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required — see backend/migrations/0001_init.sql and ADR-0012")
	}
	db, err := experiment.OpenDB(dsn)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer db.Close()
	repo := experiment.NewCachedRepository(experiment.NewPostgresRepository(db), 3*time.Second)
	candleRepo := market.NewCachedCandleRepository(market.NewPostgresCandleRepository(db), 128, time.Minute)
	jobQueue := experiment.NewPostgresQueue(db)
	workerCount := envInt("BACKTEST_WORKERS", 3, 1, 64)
	sentimentRepo := sentiment.NewPostgresRepository(db)
	sentimentURL := os.Getenv("SENTIMENT_SERVICE_URL")
	if sentimentURL == "" {
		sentimentURL = "http://localhost:8000"
	}
	sentimentClient := sentiment.NewClient(sentimentURL, &http.Client{Timeout: 3 * time.Second})
	sentimentService := sentiment.NewService(sentimentClient, sentimentRepo)
	sentimentLookup := sentiment.NewTimeLookup(sentimentRepo, sentiment.DefaultMaxAge)
	mustRegisterPlugin(registry, strategy.NewSentimentPlugin(sentimentLookup, 0.7))
	binance := market.NewBinance(nil)
	jwtSecret := os.Getenv("JWT_SECRET")
	authService, err := auth.NewService(auth.NewPostgresRepository(db), jwtSecret)
	if err != nil {
		log.Fatalf("configure auth: %v", err)
	}

	router := httpx.NewRouterWithContext(ctx, registry, repo, httpx.Dependencies{
		Auth: authService, Candles: candleRepo, Live: binance, Sentiment: sentimentService, SentimentReader: sentimentRepo,
		Queue: jobQueue, WorkerCount: workerCount, Readiness: db.PingContext,
	})
	defer router.Close()

	log.Println("listening on :8080")
	server := &http.Server{Addr: ":8080", Handler: router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envInt(name string, fallback, min, max int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		log.Fatalf("%s must be an integer between %d and %d", name, min, max)
	}
	return value
}

func mustRegisterPlugin(registry *strategy.Registry, plugin strategy.Plugin) {
	if err := registry.RegisterPlugin(plugin); err != nil {
		log.Fatal(err)
	}
}

func loadLocalEnv() {
	loadEnvFiles(".env", "backend/.env")
}

func loadEnvFiles(paths ...string) {
	for _, path := range paths {
		err := godotenv.Load(path)
		if err == nil {
			return
		}
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("load %s: %v", path, err)
			return
		}
	}
}
