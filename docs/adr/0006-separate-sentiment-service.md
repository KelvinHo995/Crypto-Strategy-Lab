# ADR-0006: Sentiment analysis is a separate Python service, not a Go package

**Status:** Accepted
**Owner:** Market Data + Sentiment Service (Võ Thành Đạt builds the News Collector and sentiment-service; Frontend consumes it in UI)

> **Implementation status (2026-09-05):** `RSSNewsProvider` implements the
> source-neutral `NewsProvider` contract, and `cmd/news-ingest` triggers the
> pipeline manually. Go upserts normalized articles into `news_items` before
> sentiment enrichment, then stores model observations in `sentiment_results`.
> Stable IDs, partial-feed failure isolation, and duplicate-analysis skipping
> are covered by deterministic tests. Scheduling and a news-read API remain
> outside the MVP.

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

The Go backend owns news collection and normalized persistence in
`news_items`. It also owns the returned sentiment observations required by
its domain, persists them separately in `sentiment_results`, and performs
bounded, time-aligned lookup for `SentimentStrategy`. This second table is a
Go-owned observation projection, not a Python model cache or a database
shared with `sentiment-service`; the Python process never connects to it.

```
RSS → News Collector (Go) → normalized NewsItem → news_items (Go Postgres)
                                              │
                                              ▼
sentiment-service (Python inference via POST /analyze)
                                              │
                                              ▼
                           sentiment_results (Go observation projection)
```

Crucially, the News Collector and the Sentiment Service are also separated
from *each other* (spec ch.28): the collector only produces `NewsItem`s from
whatever source (RSS/API/crawler), and doesn't know a sentiment model
exists; the sentiment service only classifies text, and doesn't know where
the text came from. Neither is coupled to the other's implementation.

Spec §28 ("News không được gắn cứng với một crawler") names exactly three
peer implementations behind one `News Provider` abstraction — **RSS, News
API, Crawler** — all returning the same standardized `NewsItem`, "nhờ đó
việc thay nguồn dữ liệu không ảnh hưởng đến các module phía sau." RSS is
listed as a first-class option, not a lesser fallback — picking it isn't
cutting a corner against the spec, it's implementing exactly what's named.
A professor's lecture note about using an LLM to interpret HTML tags and
cache the result applies specifically to the **Crawler** branch (raw HTML
has no stable structure, so a tag-extraction step can break and need
"healing") — it isn't a requirement across all three options, since RSS
and News API both already return structured data with no tags to extract
in the first place.

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
