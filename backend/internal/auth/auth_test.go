package auth_test

import (
	"context"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/auth"
)

func TestRegisterLoginVerify(t *testing.T) {
	svc, err := auth.NewService(auth.NewMemoryRepository(), "01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	u, err := svc.Register(context.Background(), "alice", "correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	if u.PasswordHash != "" {
		t.Fatal("password hash must not serialize/expose through returned user")
	}
	token, err := svc.Login(context.Background(), "alice", "correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := svc.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Sub != u.ID {
		t.Fatalf("sub=%s want %s", claims.Sub, u.ID)
	}
	if _, err := svc.Login(context.Background(), "alice", "wrong-password"); err == nil {
		t.Fatal("wrong password accepted")
	}
}
