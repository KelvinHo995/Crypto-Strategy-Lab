package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameTaken      = errors.New("username already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	CreatedAt    int64  `json:"createdAt"`
}
type Repository interface {
	Create(context.Context, User) error
	ByUsername(context.Context, string) (User, error)
}
type Claims struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

type Service struct {
	repo   Repository
	secret []byte
	now    func() time.Time
}

func NewService(repo Repository, secret string) (*Service, error) {
	if repo == nil {
		return nil, errors.New("auth repository is required")
	}
	if len(secret) < 16 {
		return nil, errors.New("JWT secret must contain at least 16 characters")
	}
	return &Service{repo: repo, secret: []byte(secret), now: time.Now}, nil
}

func (s *Service) Register(ctx context.Context, username, password string) (User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if len(username) < 3 || len(username) > 50 {
		return User{}, errors.New("username must contain 3-50 characters")
	}
	if len(password) < 8 || len([]byte(password)) > 72 {
		return User{}, errors.New("password must contain 8-72 bytes")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return User{}, fmt.Errorf("generate user id: %w", err)
	}
	user := User{ID: base64.RawURLEncoding.EncodeToString(idBytes), Username: username, PasswordHash: string(hash), CreatedAt: s.now().UnixMilli()}
	if err := s.repo.Create(ctx, user); err != nil {
		return User{}, err
	}
	user.PasswordHash = ""
	return user, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.ByUsername(ctx, strings.ToLower(strings.TrimSpace(username)))
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	now := s.now()
	return s.sign(Claims{Sub: user.ID, Iat: now.Unix(), Exp: now.Add(time.Hour).Unix()})
}

func (s *Service) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	unsigned := parts[0] + "." + parts[1]
	want := hmac.New(sha256.New, s.secret)
	_, _ = want.Write([]byte(unsigned))
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(sig, want.Sum(nil)) {
		return Claims{}, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}
	if json.Unmarshal(headerBytes, &header) != nil || header.Algorithm != "HS256" || header.Type != "JWT" {
		return Claims{}, ErrInvalidToken
	}
	var claims Claims
	now := s.now().Unix()
	if json.Unmarshal(payload, &claims) != nil || claims.Sub == "" || claims.Iat <= 0 || claims.Iat > now+60 || claims.Exp <= now || claims.Exp <= claims.Iat {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}

func (s *Service) sign(claims Claims) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
