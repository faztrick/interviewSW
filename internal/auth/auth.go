package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")

type Service struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

type Claims struct {
	Sub int64 `json:"sub"`
	Exp int64 `json:"exp"`
	Iat int64 `json:"iat"`
	Iss string `json:"iss"`
}

func NewService(secret, issuer string, ttl time.Duration) *Service {
	return &Service{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}
}

func HashPassword(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func VerifyPassword(raw, hashed string) bool {
	return hmac.Equal([]byte(HashPassword(raw)), []byte(hashed))
}

func (s *Service) GenerateToken(userID int64) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Sub: userID,
		Exp: now.Add(s.ttl).Unix(),
		Iat: now.Unix(),
		Iss: s.issuer,
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal jwt claims: %w", err)
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerBytes)
	claimsPart := base64.RawURLEncoding.EncodeToString(claimsBytes)
	payload := headerPart + "." + claimsPart

	sig := sign(payload, s.secret)
	return payload + "." + sig, nil
}

func (s *Service) ParseToken(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expected := sign(unsigned, s.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return Claims{}, ErrInvalidToken
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	var claims Claims
	if err = json.Unmarshal(claimsBytes, &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}

	now := time.Now().UTC().Unix()
	if claims.Exp <= now || claims.Iss != s.issuer || claims.Sub == 0 {
		return Claims{}, ErrInvalidToken
	}

	return claims, nil
}

func sign(input string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
