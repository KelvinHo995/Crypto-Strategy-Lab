package httpx

import "net/http"

// TODO: needs a WS upgrade mechanism (stdlib has none) — decide a library
// separately. Carries CANDLE_UPDATE/SEARCH_PROGRESS/LEADERBOARD_UPDATED (ADR-0008).

func serveWebSocket(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
