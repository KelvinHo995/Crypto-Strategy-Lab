# Fullstack E2E test runbook

This runbook verifies the real PostgreSQL/Supabase, Binance REST/WebSocket,
Go API, React client and Python sentiment path. Generated market data is valid
only when the user explicitly enters offline `DEMO` mode.

## 1. Prerequisites

- PostgreSQL or Supabase is reachable.
- Binance REST and WebSocket endpoints are reachable.
- Go, pnpm and Python 3.11+ are installed.
- `backend/.env` exists and is not committed:

```dotenv
DATABASE_URL=postgresql://...
JWT_SECRET=replace-with-at-least-16-characters
SENTIMENT_SERVICE_URL=http://127.0.0.1:8000
```

Apply every SQL file in `backend/migrations/` in numeric order (`0001` through
`0007`) with:

```powershell
cd backend
go run ./cmd/migrate
```

The runner uses `DATABASE_URL`; re-running is safe because migrations are
idempotent. `0005` adds the leaderboard index, `0006` creates the durable
experiment job queue, and `0007` normalizes search metadata.

## 2. Run automated regression

From three terminals:

```powershell
cd backend
go test -count=3 ./...
go vet ./...
```

The durable-queue runtime test is deliberately opt-in because a running backend
would compete for its real Supabase jobs. Stop the backend, then run:

```powershell
$env:RUN_POSTGRES_QUEUE_INTEGRATION='1'
go test -count=3 ./internal/experiment -run TestPostgresQueue
Remove-Item Env:RUN_POSTGRES_QUEUE_INTEGRATION
```

```powershell
cd frontend
pnpm install --frozen-lockfile
pnpm run lint
pnpm run build
```

```powershell
cd sentiment-service
.\.venv\Scripts\python.exe -m unittest discover -s tests -v
```

All commands must exit with code 0 before continuing.

## 3. Backfill real candles

The following command fills 180 recent days for the complete catalog. It is
idempotent but can take several minutes on a new database.

```powershell
cd backend
$env:BACKFILL_SYMBOLS='BTCUSDT,ETHUSDT,BNBUSDT,SOLUSDT,XRPUSDT,ADAUSDT,DOGEUSDT,AVAXUSDT'
$env:BACKFILL_DAYS='180'
$env:BACKFILL_TIMEFRAMES='5m,15m,1h,4h'
$env:BACKFILL_REQUEST_PAUSE_MS='175'
go run ./cmd/backfill
```

Every symbol/timeframe must report a non-zero count. Refresh only the latest
two days before a demo by changing `BACKFILL_DAYS` to `2` and rerunning.

## 4. Start the stack

Keep each process in its own terminal:

```powershell
cd sentiment-service
.\.venv\Scripts\python.exe -m uvicorn app.main:app --reload --port 8000
```

```powershell
cd backend
go run ./cmd/server
```

```powershell
cd frontend
pnpm dev
```

Expected health checks:

- `http://127.0.0.1:8000/health` returns the sentiment service health response.
- `http://127.0.0.1:8080/health` returns the Go service health response.
- `http://127.0.0.1:5173/` opens the React application.

## 5. Browser acceptance scenarios

1. Create a new account, sign in and confirm the header says `Live infrastructure`.
2. On **Market**, confirm four charts show `LIVE/API`, connection state is
   `WS Connected`, and Binance Trades continues adding rows.
3. Open every chart selector and confirm BTC, ETH, BNB, SOL, XRP, ADA, DOGE and
   AVAX are present. Switch at least one chart and the trade feed to another
   symbol; their data must change without a page reload.
4. On **Strategies**, choose a non-BTC market, enable MA + RSI and start the
   search. The request should start immediately and eventually update the
   leaderboard.
5. On **Backtests**, confirm the default date range is the latest 90 days, run
   a simulation and wait for `COMPLETED`. Open its result/provenance details.
6. On **Market news**, run the sample sentiment action. It must be labelled
   `LIVE`; the fixture article feed and aggregate remain labelled `DEMO`.
7. Reload the page. The httpOnly session cookie should keep the user signed in
   and the WebSocket should reconnect and restore active subscriptions.

Any `MOCK` market candle/trade label while signed into `Live infrastructure` is
a failure. An explicit API error is expected when a requested range was not
backfilled; the client must not disguise it with generated data.

## 6. Authenticated API smoke test

Run this while all three services are running. Change the username for each new
database. The script verifies auth, the eight-market catalog, recent candles and
one asynchronous ADA search.

```powershell
$base = 'http://127.0.0.1:8080'
$username = 'e2e-user-001'
$password = 'E2E-password-2026!'
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$credentials = @{username=$username;password=$password} | ConvertTo-Json
try {
  Invoke-WebRequest -Uri "$base/auth/register" -Method Post -ContentType 'application/json' -Body $credentials | Out-Null
} catch {
  if ($_.Exception.Response.StatusCode.value__ -ne 409) { throw }
}
Invoke-WebRequest -Uri "$base/auth/login" -Method Post -ContentType 'application/json' -Body $credentials -WebSession $session | Out-Null

$markets = Invoke-RestMethod -Uri "$base/markets" -WebSession $session
if (@($markets).Count -ne 8) { throw "Expected 8 markets" }

$to = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
$from = $to - (90L * 24 * 60 * 60 * 1000)
foreach ($market in $markets) {
  $candles = Invoke-RestMethod -Uri "$base/candles?symbol=$($market.symbol)&timeframe=1h&from=$from&to=$to&limit=500" -WebSession $session
  if (@($candles).Count -lt 21) { throw "Insufficient candles for $($market.symbol)" }
}

$request = @{pair='ADAUSDT';timeframe='1h';from=$from;to=$to;capital=10000;strategies=@('MA','RSI')} | ConvertTo-Json
$started = Invoke-RestMethod -Uri "$base/search/start" -Method Post -ContentType 'application/json' -Body $request -WebSession $session
for ($attempt = 0; $attempt -lt 40; $attempt++) {
  $result = Invoke-RestMethod -Uri "$base/experiments/$($started.searchId)" -WebSession $session
  if ($result.status -in @('COMPLETED','FAILED')) { break }
  Start-Sleep -Milliseconds 250
}
if ($result.status -ne 'COMPLETED') { throw "Search ended as $($result.status)" }
$result | Select-Object id,status,return,winRate,tradeCount,totalProfit
```

Expected negative checks:

- Calling `/markets` without the session cookie returns HTTP 401.
- Starting a search with an unsupported pair returns HTTP 400.
- Starting a search over a range with fewer than 21 persisted candles returns
  HTTP 422 and never substitutes fixture candles.

## 7. Evidence to record

For a release/demo, record the commit hash, automated-test output, the eight
candle counts, one completed experiment ID, the sentiment response model/version
and a Market screenshot showing `LIVE/API` plus the live trade feed. Never save
session cookies, `.env` contents or database credentials in Git.
