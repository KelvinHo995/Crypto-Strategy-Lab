package httpx

import "net/http"

// NewRouter wires public routes directly and mounts everything else behind
// requireAuth on a single catch-all — one gate, not a per-route list.
func NewRouter() http.Handler {
	public := http.NewServeMux()
	public.HandleFunc("GET /health", health)
	public.HandleFunc("POST /auth/register", register)
	public.HandleFunc("POST /auth/login", login)

	protected := http.NewServeMux()
	protected.HandleFunc("POST /auth/logout", logout)
	protected.HandleFunc("POST /search/start", startSearch)
	protected.HandleFunc("GET /experiments", listExperiments)
	protected.HandleFunc("GET /experiments/{id}", getExperiment)
	protected.HandleFunc("GET /strategies", listStrategies)
	protected.HandleFunc("GET /ws", serveWebSocket)

	public.Handle("/", requireAuth(protected))
	return public
}

// requireAuth is a stub — real JWT verification lands with internal/auth (ADR-0007).
func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
