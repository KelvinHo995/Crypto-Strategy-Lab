package httpx_test

import (
	"bytes"
	"context"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/auth"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthProtectsRoutesAndLoginSetsCookie(t *testing.T) {
	svc, _ := auth.NewService(auth.NewMemoryRepository(), "ronaldothuamessi")
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Auth: svc})
	defer router.Close()
	srv := httptest.NewServer(router)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/strategies")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauth status=%d", resp.StatusCode)
	}
	registerBody := `{"username":"alice","password":"correct-horse"}`
	resp, err = http.Post(srv.URL+"/auth/register", "application/json", strings.NewReader(registerBody))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register=%d", resp.StatusCode)
	}
	resp, err = http.Post(srv.URL+"/auth/login", "application/json", bytes.NewBufferString(registerBody))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login=%d", resp.StatusCode)
	}
	var session *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session" {
			session = c
		}
	}
	if session == nil || !session.HttpOnly {
		t.Fatal("secure session cookie missing")
	}
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/strategies", nil)
	req.AddCookie(session)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("authenticated status=%d", resp.StatusCode)
	}
}
