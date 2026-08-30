package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/auth"
)

const sessionCookie = "session"

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func register(svc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in credentials
		if decodeJSON(w, r, &in) != nil {
			http.Error(w, "invalid request body", 400)
			return
		}
		u, err := svc.Register(r.Context(), in.Username, in.Password)
		if errors.Is(err, auth.ErrUsernameTaken) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(u)
	}
}
func login(svc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in credentials
		if decodeJSON(w, r, &in) != nil {
			http.Error(w, "invalid request body", 400)
			return
		}
		token, err := svc.Login(r.Context(), in.Username, in.Password)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: 3600, Expires: time.Now().Add(time.Hour)})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "AUTHENTICATED"})
	}
}
func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
	w.WriteHeader(http.StatusNoContent)
}

func requireAuth(svc *auth.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if svc == nil {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		if _, err = svc.Verify(cookie.Value); err != nil {
			http.Error(w, "invalid or expired session", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
