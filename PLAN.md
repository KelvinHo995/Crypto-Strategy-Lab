# Kế hoạch Build Crypto Strategy Lab — Nhóm 4 người, 2 tuần

Stack: **Go 1.22+** (backend core, `net/http` thuần — không dùng chi, xem mục 6), **React + TypeScript** (npm), **Python 3.11 + FastAPI** (Sentiment Service, quản lý bằng `uv`), **PostgreSQL (Supabase-hosted, dùng chung cho cả nhóm — xem ADR-0012)**

Repo: monorepo, `backend/` · `frontend/` · `sentiment-service/` · `docs/adr/`

---

## 0. Trạng thái hiện tại (cập nhật liên tục)

- ✅ Repo scaffold xong (backend/frontend/sentiment-service), README từng phần đã có
- ✅ `Candle` struct chốt xong, có thêm field `IsClosed` (candle đang hình thành vs đã đóng)
- ✅ Binance live stream đã merge và mở hai combined WebSocket dùng chung: kline cho 8 coin × 4 timeframe và aggregate trades cho 8 coin. `/ws` lọc theo subscription của từng browser, có reconnect/backoff.
- ✅ sentiment-service FastAPI chạy được; `/analyze` dùng model lexicon xác định `crypto-lexicon/v2`, validate input, trả score/model version/timestamp thật và có test. News ingestion phân tích cả headline lẫn RSS summary để tránh bỏ mất market-moving verb. FinBERT vẫn là nâng cấp ngoài MVP.
- ✅ **News Collector** — `internal/news.NewsProvider` có implementation RSS thật và `cmd/news-ingest` là trigger thủ công. Pipeline chuẩn hóa rồi upsert `news_items` trước khi gọi sentiment-service; kết quả inference được lưu riêng trong `sentiment_results`. Stable ID, partial-feed failure và skip sentiment đã có test; scheduler/API đọc news nằm ngoài MVP.
- ✅ `internal/experiment`: Backtester + Evaluator, SL/TP, gap fill, fee, slippage, tránh lookahead bias và same-candle re-entry đã có test; `go vet`/`go test ./...` sạch. `NewJobID` nay ghép UUID sau timestamp (`exp-<unixnano>-<uuid>`) nên không còn phụ thuộc độ phân giải clock của OS — an toàn cả trong vòng lặp sinh candidate liên tục lẫn giữa nhiều backend instance dùng chung `PostgresQueue`. Đã sửa lỗi thật: `Config.Window` từng cố định (20), nhỏ hơn nhu cầu thật của MA mặc định (50) và của `RandomGenerator` (tới 200) — MA lặng lẽ trả `Hold` mãi mãi, `COMPLETED` với 0 trade trông như kết quả hợp lệ. Nay mỗi strategy tự báo `MinLookback()`, worker tính window đúng theo từng candidate đã resolve — verified sống: loop với `maLongWindow` tới 192 đều có trade thật (195–3200), không còn 0. Xem ADR-0003, ADR-0009, [06-search-backtest-flow.md](docs/architecture/06-search-backtest-flow.md).
- ✅ `internal/httpx`: router 2-mux public/protected; search/experiment/strategy APIs chạy thật. `/health` là liveness, `/ready` kiểm tra PostgreSQL, `/metrics` có queue depth/running/failed, throughput và queue/execution latency; response có `X-Request-ID` đồng nhất với request-duration log. Auth dùng JWT cookie thật; `/ws` là server-push channel có xác thực cho candle, tiến độ search và leaderboard.
- ✅ `Binance.FetchHistoricalCandles` REST thật: pagination 1000 klines, normalize `Candle`, chỉ nhận candle đã đóng; live kline WebSocket có reconnect/backoff và phát `CANDLE_UPDATE`.
- ✅ Postgres repositories cho experiments/users/candles/sentiment và `cmd/backfill` đã hoàn thiện; default 2 năm/BTC, hỗ trợ `BACKFILL_SYMBOLS` + `BACKFILL_DAYS`, pace Binance request, upsert batch 500 và idempotent theo `(symbol,timeframe,open_time)`. Search production đọc range qua bounded cache. Chuỗi migration hiện chạy idempotent từ `0001` đến `0014`: durable queue/search metadata, strategy instances, article metadata, trade detail, normalized news, sentiment-model provenance, pair/timeframe provenance và conservative legacy backfill.
- ✅ Strategy thật (MA/RSI/BB/SR/SMC/MACD/Sentiment), `Registry.Get/List`, `CombinationPolicy`, `StrategyGenerator` — đã hoàn thiện. `RegisterPlugin(Plugin)` đăng ký atomically implementation, factory, random parameter generator và strategy version; architecture tests chứng minh plugin mới chạy xuyên generate→composition→backtest→evaluate→rank mà không sửa core consumers. `CandidateStrategy.Instances []StrategyInstance` cho mỗi instance tự có `type/params/weight`; `resolveCombinationPolicy` dùng đúng weight client gửi (fallback equal nếu không set). Migration `0008` lưu instances JSON. Frontend Composite Strategy Builder gửi contract này thật — verified sống qua Supabase với 2 instance MA khác param/weight.
- ✅ Queue/Worker pool dùng durable PostgreSQL queue (`BACKTEST_WORKERS`, mặc định 3, giới hạn 1–64), claim bằng `SKIP LOCKED`, retry/backoff, lease heartbeat và reclaim sau crash. `cmd/perf` đo cùng workload với 1/3 workers; runtime metrics đo throughput, queue wait, execution time và failure count mà không đổi backtest/search semantics. `/search/start` và `/search/loop` ghi `PENDING + job` atomically; worker đọc candle qua bounded cache và cô lập panic.
- ⚠️ `POST /search/loop` (ADR-0011) đã có `RandomGenerator`, nhận max-candidates / max-duration / no-improvement, và job ID nay an toàn (UUID). Backend phát terminal `SEARCH_PROGRESS{searchId,status:STOPPED|FAILED,reason}` ngay khi loop dừng sớm (max-duration/no-improvement/cancel/enqueue-failed), nên progress không còn treo dưới `total` vô thời hạn — verified sống qua Supabase thật. Còn tồn hai giới hạn đã biết, chấp nhận được ở quy mô hiện tại: PostgreSQL queue không backpressure nên no-improvement có thể overshoot nhiều hơn "vài candidate" như comment cũ giả định; dedup retry tối đa 10 lần rồi vẫn có thể nhận duplicate. User-cancel vẫn chưa có endpoint (3 điều kiện còn lại đã đảm bảo dừng).
- ✅ Frontend Discovery đã gọi thật `POST /search/loop`, gửi đủ ba giới hạn, bỏ timer/progress giả và dùng `SEARCH_PROGRESS` + `LEADERBOARD_UPDATE` trong LIVE mode. UI hỗ trợ `COMPLETED|STOPPED|FAILED`; backend nay phát đủ `searchId/status/reason` cho `STOPPED/FAILED` cấp search-run, `COMPLETED` vẫn suy ra từ `tested == total`. Auth, catalog 8 coin, candles, aggregate trades, strategy list, experiments và sentiment analyze vẫn nối API thật; phần news aggregate trên frontend vẫn có DEMO fallback.
- ✅ **Auth (users/session)** — bcrypt, Postgres user repository, JWT HS256 1h, httpOnly SameSite=Lax cookie, protected-route middleware và logout cookie clearing đã hoàn thiện; `JWT_SECRET` tối thiểu 16 ký tự là biến môi trường bắt buộc.

## 1. Phân công (đã điều chỉnh so với bản đầu)

Đổi so với kế hoạch gốc: Người viết file này chuyển từ Market Data sang **Experiment** (domain nặng nhất), Người 1 gánh **cả Market Data lẫn Sentiment**, Người 2 giữ nguyên Strategy (domain nhẹ nhất, theo yêu cầu).

| | Bạn (Experiment) | Người 1 | Người 2 | Người 4 |
|---|---|---|---|---|
| **Domain chính** | Experiment: Backtest/Evaluate/Rank | Market Data (Go) | Strategy + Search (Go) | Frontend (React) |
| **Domain phụ** | Job Queue/Worker (M5, chủ trì) | Sentiment Service (Python, sau khi xong Market Data) | Hỗ trợ Job Queue/Worker khi cần | — |
| **Vì sao** | Domain nặng nhất: correctness cao, provenance, transaction boundary, là điểm hội tụ mọi domain khác | Việc dồn đầu (WS+REST+DB), nhẹ dần, rồi nặng lại với Python ở cuối | Nhẹ nhất — công thức strategy, không I/O phức tạp | Việc trải đều liên tục suốt 14 ngày, không dồn cục |

**Quan trọng:** Job Queue/Worker (M5) là phần khó thứ nhì sau Experiment — **không để mặc định rơi vào Người 2** chỉ vì nó nằm chung milestone "Search". Chủ trì là bạn (Experiment), Người 2 hỗ trợ khi rảnh, Người 1 đứng dự phòng nếu cần sau khi xong Market Data.

---

## 2. Lịch 14 ngày

| Ngày | Người 1 | Người 2 | Người 3 | Người 4 |
|---|---|---|---|---|
| 1 | Họp chung: chốt contract (Candle, Strategy interface, ExperimentResult, Sentiment API, WS message) — **bắt buộc xong trong ngày**, không thì cả nhóm code lệch pha | | | |
| 2–3 | Binance Adapter → Candle chuẩn | ✅ 5 strategy: MA/RSI/BB/SR + SMC (SMC bản tối giản — swing high/low structure break, không cần đúng 100% lý thuyết SMC, xem PDF ch.11) (chạy với mock data, không chờ Người 1) | Backtester + Evaluator (mock signal, gồm SL/TP/transaction cost/slippage 5bps) | React skeleton + chart component (mock WS data) |
| 4–5 | ✅ Nối Binance WebSocket thật → Backend; reconnect/backoff | ✅ StrategyRegistry.register() + extension test | Nối signal thật từ Người 3; bắt đầu transaction boundary | Nối chart vào WS thật; UI chọn strategy |
| 6 | Bắt đầu Sentiment Service (FastAPI) | ✅ CandidateStrategy generator (Random) | Experiment pipeline: Candidate→Backtest→Evaluate→Rank + provenance field | Leaderboard UI |
| 7 | Sentiment API hoàn chỉnh + test | Nhảy sang hỗ trợ Người 3: Job Queue/Worker pool | ✅ Job Queue/Worker pool (cùng Người 2) | UI search progress / observability panel |
| 8–9 | **Buffer chung — fix bug tích hợp toàn hệ thống** | | | |
| 10 | ✅ Go sentiment REST client + lỗi service-down rõ ràng | ✅ SentimentStrategy fallback về base strategy khi service lỗi | ✅ Experiment snapshot `strategyVersions`; sentiment client trả model name/version thật | News panel UI + gắn Sentiment vào chart/leaderboard |
| 11 | Test failure case: News/Sentiment down | ✅ Đo throughput 1→3 bằng `cmd/perf` | ✅ Worker count là env scaling knob; `/metrics` đo production queue/runtime | Đảm bảo FE không sập khi News down |
| 12 | **Architecture Proof cả nhóm** — mỗi người test domain mình (Extensibility / Replaceability / Scalability & Failure) | | | |
| 13 | Viết ADR (5–6 cái quan trọng), chuẩn bị data demo | | | |
| 14 | Rehearsal demo, chuẩn bị trả lời checklist vấn đáp | | | |

**Ưu tiên cắt nếu thiếu giờ** (theo thứ tự): Event Sourcing/CQRS thật (dùng CRUD đơn giản) → Genetic Search (giữ Random, chứng minh bằng interface) → MLOps monitoring dashboard riêng (chỉ cần lưu model/version trong record) → Serverless cho News (cron job thường cũng được).

**Ngoài phạm vi 2 tuần (stretch, không lên lịch trong bảng trên)**: NL/link-to-strategy generator (nhập ngôn ngữ tự nhiên hoặc link website, hệ thống tự sinh strategy) — chi phí implement lớn (LLM synthesis + validate output khớp `Strategy` interface + UI riêng), không chứng minh thêm gì về kiến trúc so với `StrategyGenerator` interface đã có (Random/Domain-guided); LLM-based HTML tag extraction + cache schema cho News Crawler — dùng parser cố định theo RSS/News API là đủ cho MVP, raw-HTML scraping không cần thiết khi có nguồn structured. Chi tiết lý do: `docs/architecture/07-quality-attributes.md`.

---

## 3. Contract / Interface dùng chung

### 3.1 Candle — giữa Market Data và mọi domain khác

```go
type Candle struct {
    Symbol    string  `json:"symbol"`
    Timeframe string  `json:"timeframe"` // "5m","15m","1h","4h"
    OpenTime  int64   `json:"openTime"`  // unix ms
    Open      float64 `json:"open"`
    High      float64 `json:"high"`
    Low       float64 `json:"low"`
    Close     float64 `json:"close"`
    Volume    float64 `json:"volume"`
    IsClosed  bool    `json:"isClosed"` // false khi candle vẫn đang hình thành
}
```

### 3.2 Strategy — giữa Strategy và Experiment (nội bộ Go)

```go
type Signal string
const (
    Buy  Signal = "BUY"
    Sell Signal = "SELL"
    Hold Signal = "HOLD"
)

type Strategy interface {
    Name() string
    Analyze(candles []Candle) Signal
}

type StrategyRegistry struct {
    strategies map[string]Strategy
}
type Plugin struct {
    Name         string
    Default      Strategy
    Factory      StrategyFactory
    RandomParams RandomParamsGenerator
    Version      string
}
func (r *StrategyRegistry) RegisterPlugin(plugin Plugin) error

type StrategyGenerator interface {
    Generate() CandidateStrategy
}

type CandidateStrategy struct {
    ID        string
    Instances []StrategyInstance
    Policy    string // "majority" | "weighted"
}
```

### 3.3 ExperimentResult — giữa Experiment và Frontend

```go
type ExperimentResult struct {
    ID               string            `json:"id"`
    CandidateID      string            `json:"candidateId"`
    Pair             string            `json:"pair"`            // empty only for legacy rows
    Timeframe        string            `json:"timeframe"`       // actual runtime candle interval
    Strategies       []string          `json:"strategies"`       // snapshot, not FK
    Params           map[string]any    `json:"params"`           // snapshot
    Policy           string            `json:"policy"`           // snapshot
    StrategyVersions map[string]string `json:"strategyVersions"` // provenance
    SentimentModels  []ModelIdentity   `json:"sentimentModels"`  // runtime identities actually consumed
    DatasetPeriod    string            `json:"datasetPeriod"`
    Return           float64           `json:"return"`
    MDD              float64           `json:"mdd"`
    TradeCount       int               `json:"tradeCount"`
    WinRate          float64           `json:"winRate"`
    Wins             int               `json:"wins"`
    Losses           int               `json:"losses"`
    TotalProfit      float64           `json:"totalProfit"`
    Status           string            `json:"status"` // PENDING|RUNNING|COMPLETED|FAILED
    CreatedAt        int64             `json:"createdAt"`
}
```

### 3.4 Sentiment API — Go ↔ Python FastAPI (REST)

```
POST http://sentiment-service:8000/analyze
Request:  { "newsId": "8821", "text": "..." }
Response: {
  "newsId": "8821",
  "sentiment": "NEGATIVE",
  "score": 0.91,
  "model": { "name": "crypto-lexicon", "version": "v2" },
  "createdAt": 1723000000
}
```

### 3.5 WebSocket — Go ↔ React

```json
{ "type": "CANDLE_UPDATE", "payload": { /* Candle, kể cả isClosed=false — tick đang chạy */ } }
{ "type": "TRADE_TICK", "payload": { /* TradeTick từ Binance aggTrade */ } }
{ "type": "LEADERBOARD_UPDATE", "payload": { /* ExperimentResult[] */ } }
{ "type": "SEARCH_PROGRESS", "payload": { "tested": 125, "total": 500 } }
```

### 3.6 Auth — User & Session (rất tối giản: không role, không reset password, không OAuth)

```go
type User struct {
    ID           string `json:"id"`
    Username     string `json:"username"`
    PasswordHash string `json:"-"`      // bcrypt, không bao giờ serialize ra JSON
    CreatedAt    int64  `json:"createdAt"`
}
```

Session: JWT ký bằng secret server-side (HS256), claims `{sub: userID, iat,
exp}`, `exp = iat + 1h` — không phải session token tra DB (lý do xem
ADR-0007). Gửi về client qua httpOnly cookie khi login; browser tự gửi lại ở
mọi request kể cả lúc WebSocket handshake. Middleware chỉ verify chữ ký +
`exp`, không query DB. Toàn bộ API — trừ `/health`, `/ready`, `POST /auth/register`,
`POST /auth/login` — yêu cầu JWT hợp lệ, chưa hết hạn. Không có danh sách
route "public vs private" riêng — 1 middleware áp dụng đều, để bớt chỗ dễ
sai. Logout chỉ xoá cookie phía client — token cũ về lý thuyết vẫn hợp lệ
tới khi hết hạn tự nhiên (tối đa 1h), đây là đánh đổi đã chấp nhận, không
phải thiếu sót.

### 3.7 Queue — BacktestJob & Job Queue interface (giữa Search Loop và Backtest Worker)

```go
type BacktestJob struct {
    ID         string
    Candidate  CandidateStrategy // xem 3.2
    EnqueuedAt int64
}

type Queue interface {
    Enqueue(ctx context.Context, job BacktestJob) error
    Dequeue(ctx context.Context) (BacktestJob, error)
}
```

`PostgresQueue` là implementation production; `InMemoryQueue` chỉ còn dùng cho
unit test/fallback composition. Worker pool và Backtester vẫn phụ thuộc interface
`Queue`, không phụ thuộc transport. Xem
[ADR-0013](docs/adr/0013-postgres-durable-queue-and-local-caches.md).

---

## 3b. Toàn bộ API endpoint (~7-8 cái, không nhiều)

### Go backend

| Method | Path | Việc |
|---|---|---|
| POST | `/search/start` | Bắt đầu search — trả về ngay `{searchId, status: STARTED}`, không block chờ chạy xong |
| POST | `/search/loop` | Bắt đầu Random Search 2–200 candidate với max-duration/no-improvement optional; trả HTTP 202 ngay |
| GET | `/experiments` | Leaderboard — competitive run (`COMPLETED`, có trade, strategy instances và market provenance) đứng trước; run legacy/no-trade vẫn giữ để audit nhưng không chiếm Top-K mặc định |
| GET | `/experiments/{id}` | Chi tiết 1 kết quả (click Top #1), gồm provenance |
| GET | `/strategies` | List authoritative plugin names; UI vẫn hiện plugin chưa có metadata local bằng default factory params |
| GET | `/markets` | Catalog 8 coin và các timeframe được backend hỗ trợ |
| GET | `/candles` | Nến lịch sử đã backfill từ Postgres |
| GET | `/ws` | WebSocket — CANDLE_UPDATE, TRADE_TICK, SEARCH_PROGRESS, LEADERBOARD_UPDATE; client gửi lệnh subscribe/unsubscribe |
| GET | `/health` | Sanity check, hữu ích cho docker-compose/demo |
| GET | `/ready` | Readiness check PostgreSQL, trả 503 nếu dependency chưa sẵn sàng |
| GET | `/metrics` | JSON queue depth/running/failed, throughput, queue wait và execution time |
| GET | `/metrics/prometheus` | Prometheus exposition cho worker/queue/failure/latency gauges |
| POST | `/auth/register` | `{username, password}` → tạo user, hash password bằng bcrypt |
| POST | `/auth/login` | `{username, password}` → set session cookie |
| POST | `/auth/logout` | Xoá session hiện tại |

**Tất cả endpoint khác ở trên** (`/search/start`, `/search/loop`, `/experiments`,
`/experiments/{id}`, `/strategies`, `/metrics`, `/ws`) **yêu cầu session hợp lệ** — trừ
`/health`, `/ready` và 2 endpoint `/auth/register` + `/auth/login`.

### sentiment-service (Python)

| Method | Path | Việc |
|---|---|---|
| POST | `/analyze` | Input text → sentiment + score + model version |
| GET | `/health` | Sanity check |

**Không có form validate phức tạp** — chỉ 2-3 request body cần check (`StartSearchRequest`, `AnalyzeRequest`), dùng pattern `Validate() error` viết tay trên struct (xem mục 6), không cần thư viện `validator`.

## 3c. DB schema — Postgres (Supabase-hosted, xem ADR-0012), dữ liệu thuộc Go backend

Lưu ý: cột lưu unix-millisecond timestamp phải là `BIGINT`, không phải
`INTEGER` — `INTEGER` trong Postgres chỉ 32-bit (tối đa ~2.1 tỷ), unix ms
hiện tại đã ~1.7 nghìn tỷ, sẽ overflow. Đây là khác biệt thật giữa SQLite
(INTEGER affinity linh hoạt tới 64-bit) và Postgres — xem ADR-0012.

```sql
-- Toàn bộ leaderboard + provenance nằm ở đây, đọc/ghi qua 1 Repository interface
CREATE TABLE experiments (
    id                 TEXT PRIMARY KEY,
    candidate_id       TEXT NOT NULL,
    pair               TEXT NOT NULL DEFAULT '', -- empty means unknown legacy provenance
    timeframe          TEXT NOT NULL DEFAULT '', -- empty means unknown legacy provenance
    strategies         TEXT NOT NULL,   -- JSON: ["MA","RSI"] — snapshot, not FK
    params             TEXT NOT NULL,   -- JSON: {"maWindow":20,...} — snapshot
    policy             TEXT NOT NULL,   -- snapshot
    strategy_versions  TEXT NOT NULL,   -- JSON provenance
    sentiment_models   TEXT NOT NULL,   -- JSON [{"name":"...","version":"..."}]
    dataset_period     TEXT NOT NULL,
    return_pct         REAL,
    mdd                REAL,
    trade_count        INTEGER,
    win_rate           REAL,
    wins               INTEGER,
    losses             INTEGER,
    total_profit       REAL,
    status             TEXT NOT NULL,
    created_at         BIGINT NOT NULL
);

-- Dữ liệu nến lịch sử, backfill 1 lần, dùng cho mọi lần backtest
CREATE TABLE candles (
    symbol     TEXT NOT NULL,
    timeframe  TEXT NOT NULL,
    open_time  BIGINT NOT NULL,
    open REAL, high REAL, low REAL, close REAL, volume REAL,
    PRIMARY KEY (symbol, timeframe, open_time)
);

-- Durable backtest work queue; payload chỉ chứa dataset reference/config,
-- không chứa mảng candle lớn.
CREATE TABLE experiment_jobs (
    id           TEXT PRIMARY KEY,
    payload      JSONB NOT NULL,
    status       TEXT NOT NULL,
    attempts     INTEGER NOT NULL DEFAULT 0,
    available_at BIGINT NOT NULL,
    locked_at    BIGINT,
    last_error   TEXT,
    created_at   BIGINT NOT NULL
);

-- Projection do Go backend sở hữu để lookup sentiment theo thời gian candle.
-- Đây không phải cache/model state của sentiment-service.
CREATE TABLE sentiment_results (
    news_id       TEXT PRIMARY KEY,
    published_at  BIGINT NOT NULL,
    sentiment     TEXT NOT NULL,
    score         REAL NOT NULL,
    model_name    TEXT NOT NULL,
    model_version TEXT NOT NULL,
    analyzed_at   BIGINT NOT NULL
);

-- Dữ liệu news đã chuẩn hóa do Go backend thu thập, độc lập với enrichment.
CREATE TABLE news_items (
    id            TEXT PRIMARY KEY,
    title         TEXT NOT NULL,
    content       TEXT NOT NULL,
    source        TEXT NOT NULL,
    url           TEXT NOT NULL,
    published_at  BIGINT NOT NULL,
    related_coins JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at    BIGINT NOT NULL
);
```

**Backfill dataset:** mặc định BTC/2 năm; operator có thể giới hạn coin và số
ngày bằng `BACKFILL_SYMBOLS`/`BACKFILL_DAYS`. Chạy thủ công hoặc định kỳ, không
chạy trong server startup; upsert theo primary key nên chạy lại an toàn.

**Không cần bảng riêng cho strategies** vì strategy là code trong
`StrategyRegistry`. `sentiment-service` sở hữu model inference và mọi
model-internal cache/state; nếu cần persistence riêng cho các concern đó thì
nó không dùng chung Postgres của Go. Go backend chỉ sở hữu
`news_items` cho news đã chuẩn hóa và `sentiment_results` cho observation
projection nhận từ API để lookup theo thời gian cho strategy. Hai bảng thuộc
Go backend; bảng thứ hai không phải cache của Python service.

```sql
-- Đăng nhập tối giản — không role, không password reset, không OAuth
-- Không có bảng sessions: session là JWT tự-chứa (self-contained), verify
-- bằng chữ ký + exp claim, không tra DB — xem ADR-0007
CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,   -- bcrypt
    created_at    BIGINT NOT NULL
);
```

Connection string tới Supabase là secret dùng chung cả nhóm — mỗi người giữ
riêng trong `.env.local` (gitignored), không paste vào chat/commit. Chỉ Go
backend kết nối Postgres trực tiếp; frontend không bao giờ gọi Supabase.

---

## 6. Các quyết định kỹ thuật đã chốt (và lý do KHÔNG dùng)

Việc biết mình *không* cần gì cũng là kiến trúc — mỗi dòng dưới đây đều có thể trả lời được trong vấn đáp bằng "không có driver nào bắt buộc cho scope này":

| Công nghệ/pattern | Quyết định | Vì sao |
|---|---|---|
| Web framework (chi, gin) | **Không** — `net/http` thuần (Go 1.22+ đã có path params + method routing sẵn) | API hiện tại vẫn đủ nhỏ cho standard-library routing và helper trong `internal/httpx`; chưa có driver cần framework |
| Redis | **Không** | Traffic thấp, 1 instance, không cần shared cache giữa nhiều replica |
| Kafka / message broker | **Không** — dùng in-process job queue (channel + worker pool) | Không có nhiều service cần decouple qua broker thật |
| CQRS / Event Sourcing | **Không** — 1 bảng `experiments`, CRUD đơn giản | Write model và read model hình dạng giống hệt nhau, dữ liệu mỏng, không có áp lực đọc |
| Kubernetes | **Không** — chạy chay lúc dev, docker-compose chỉ cho demo cuối | Không cần orchestration cho 1 server + 1 FE + 1 sentiment service |
| Request validator library (Go) | **Không** — viết tay `Validate() error` trên struct | Chỉ 2-3 struct cần validate, học 1 thư viện tốn công hơn viết tay |
| `var _ Interface = (*Type)(nil)` | **Có dùng** — mỗi Strategy implementation | Ép compiler bắt lỗi lệch signature ngay tại chỗ, miễn phí |
| Dependency injection qua interface (Repository) | **Có dùng** | Cho phép bắt đầu code với in-memory store, swap sang Postgres (Supabase) sau không đụng handler — xem ADR-0012 |
| uv (Python), pnpm (Node) | **Có dùng** | Nhanh hơn pip/npm, lockfile thật, cài đặt nhất quán giữa các máy |

## 6b. Quy ước Git

- Nhánh: `feat/<domain>-<slice-cụ-thể>` — VD: `feat/market-data-websocket`, `feat/market-data-backfill` — **không** làm 1 nhánh to sống cả tuần theo domain, mỗi slice việc xong là PR + merge ngay, giữ nhánh ngắn hạn (1-2 ngày)
- 1 approval trước khi merge `main`
- Người phụ trách domain nào review PR đụng vào domain đó

---

## 4. ADR cần viết (rút gọn 5–6 cái quan trọng nhất)

Format: Context · Decision · Alternatives · Consequences · Evidence

1. **Why MarketDataProvider + Adapter?** (Người 1)
2. **Why Strategy Plugin/Registry?** (Người 2)
3. **Why separate Backtester and Evaluator?** (Người 3)
4. **Why queue/worker (or why NOT Kafka)?** (Người 2 + 3)
5. **Why modular monolith vs microservices?** (cả nhóm)
6. **Why separate News Collector and Sentiment Service?** (Người 1 — Võ Thành Đạt; ownership và pipeline hiện tại được ghi trong ADR-0006)
7. **Why JWT (1h expiry) instead of a server-side session table?** (cần gán người — xem §0)

---

## 5. Chuẩn bị vấn đáp — 10 câu checklist, ai trả lời câu nào

| # | Câu hỏi | Người phụ trách trả lời |
|---|---|---|
| 1 | Architectural drivers là gì? | Cả nhóm (thống nhất chung) |
| 2 | C4 Context và Container của nhóm? | Cả nhóm |
| 3 | Boundary của Market / Strategy / Experiment / News? | Người tương ứng từng domain |
| 4 | Thêm strategy mới sửa ở đâu? | Người 2 |
| 5 | Đổi search algorithm sửa ở đâu? | Người 2 |
| 6 | Provider mới có làm frontend đổi không? | Người 1 + Người 4 |
| 7 | 100.000 backtests scale thế nào? | Người 3 (+ Người 2 hỗ trợ Queue) |
| 8 | Service lỗi có lan failure không? | Người 4 (test News down) + Người 1 (test Binance disconnect) |
| 9 | Duplicate/retry/event order xử lý thế nào? | Người 3 |
| 10 | Leaderboard result truy được provenance thế nào? | Người 3 |
| 11 | Vì sao chọn JWT 1h thay vì session table? Logout thật sự nghĩa là gì (và vì sao chấp nhận được)? | (cần gán người — xem §0) |

**Lưu ý:** Không trả lời được câu nào trong domain mình phụ trách là dấu hiệu kiến trúc "vẫn đang là hộp và mũi tên" — mỗi người nên tự test lại domain mình bằng 3 bài Architecture Proof (Extensibility / Replaceability / Scalability & Failure) trước ngày vấn đáp, không chỉ code chạy được mà phải giải thích được *tại sao* nó chịu được thay đổi.

---

*File này là kế hoạch tự tổ chức của nhóm, không phải trích dẫn từ slide gốc [S]/[P]/[R#]/[W#].*
