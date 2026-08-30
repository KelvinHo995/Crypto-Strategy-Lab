package auth_test

import (
	"context"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/auth"
)

func TestRegisterLoginVerify(t *testing.T) {
	svc, err := auth.NewService(auth.NewMemoryRepository(), "ronaldothuamessi")
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
	if _, err := svc.Login(context.Background(), "ALICE", "correct-horse"); err != nil {
		t.Fatal("username lookup must be case-insensitive")
	}
	if _, err := svc.Login(context.Background(), "alice", "wrong-password"); err == nil {
		t.Fatal("wrong password accepted")
	}
}

func TestRejectsSecretShorterThanTeamConfiguration(t *testing.T) {
	if _, err := auth.NewService(auth.NewMemoryRepository(), "too-short"); err == nil {
		t.Fatal("short JWT secret accepted")
	}
}
