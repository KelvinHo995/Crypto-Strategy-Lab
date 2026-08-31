# ADR-0006: Sentiment analysis is a separate Python service, not a Go package

**Status:** Accepted
**Owner:** Market Data / Frontend (Person 1 builds it, Person 4 consumes it in UI)

## Context

Every other domain (Market, Strategy, Experiment) lives inside the Go
monolith ([ADR-0005](0005-modular-monolith-not-microservices.md)). Sentiment
classification is the one exception under consideration. It needs an NLP/ML
model (e.g. FinBERT-class model), and Go's ML ecosystem is far behind
Python's for this — this is a real technical constraint, not just a
preference.

## Decision

`sentiment-service` is a standalone FastAPI (Python) process with its own
`/analyze` and `/health` endpoints, reached from the Go backend over REST.
It owns model inference and any model-internal cache or state. If those
concerns ever need persistence, the service uses its own datastore rather
than sharing the Go backend's Postgres instance. It can be deployed,
restarted, or scaled independently of the Go backend.

The Go backend owns the returned sentiment observations required by its
domain. It persists them in `sentiment_results` and performs bounded,
time-aligned lookup for `SentimentStrategy`. This table is a Go-owned
observation projection, not a Python model cache or a database shared with
`sentiment-service`; the Python process never connects to it.

```
News Collector (Go, inside internal/*)
     │  collect news → NewsItem
     ▼
sentiment-service (Python, separate process)
     │  POST /analyze { newsId, text } → { sentiment, score, model{name,version} }
     ▼
Go backend — stores sentiment alongside News, optionally feeds SentimentStrategy
```

Crucially, the News Collector and the Sentiment Service are also separated
from *each other* (spec ch.28): the collector only produces `NewsItem`s from
whatever source (RSS/API/crawler), and doesn't know a sentiment model
exists; the sentiment service only classifies text, and doesn't know where
the text came from. Neither is coupled to the other's implementation.

## Alternatives considered

- **Call a Python ML library from Go via cgo/subprocess.** Rejected —
  fragile packaging, doesn't get Python's ecosystem (HuggingFace/transformers)
  cleanly, and couples Go binary builds to a Python runtime being present.
- **Rewrite the model in Go or use a Go-native sentiment library.** Rejected
  for MVP — Go's NLP tooling is far less mature; not worth the accuracy cost
  for a 2-week scope where the MVP uses a deterministic versioned lexicon model to prove the
  architecture, not to be state-of-the-art (spec ch.47: the goal is proving
  the pipeline shape, not a maximally accurate model).
- **Crawler calls the model directly, inline** (`Crawler → BERT model`).
  Explicitly listed as an anti-pattern in spec ch.44 — couples data
  collection to a specific ML implementation, meaning changing the model
  would require touching crawler code too.

## Consequences

- **Positive:** the Go backend and frontend must be able to survive
  `sentiment-service` being down or slow — charts and non-sentiment
  strategies keep working (spec ch.40 Q5: "Nếu News Service bị lỗi thì Chart
  có còn chạy không?" — this ADR's decision means the answer must be yes,
  since sentiment is not in the request path for market data or non-sentiment
  strategies).
- **Positive:** model swaps (new FinBERT version, different model entirely)
  only touch `sentiment-service` and are visible to the rest of the system
  only as a version string in `model.version` — this is what keeps
  `StrategyVersions` provenance meaningful (spec ch.36) even when the model
  changes.
- **Cost:** one more process to run for local dev and demo, one more network
  hop (Go → Python REST call) with its own latency/failure mode that must be
  explicitly handled (timeout + graceful degradation), not assumed away.
- **Implemented degradation:** the Go REST client returns an explicit error
  for unavailable/non-200 responses; `SentimentStrategy` falls back to its
  base technical strategy (or `HOLD` without one). It does not silently use
  a stale score or retry a deterministic request.
