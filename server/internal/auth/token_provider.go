package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenProvider 는 AT(JWT) 와 RT(불투명 토큰) 를 만들고 확인한다.
// main 에서 한 번 만들어 필요한 곳에 넘겨 쓴다.
type TokenProvider struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenProvider(secret string, accessTTL, refreshTTL time.Duration) *TokenProvider {
	return &TokenProvider{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// IssueAccessToken 은 accountID 를 담은 짧은 수명의 JWT 를 만든다.
func (p *TokenProvider) IssueAccessToken(accountID int64) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(accountID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(p.accessTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(p.secret)
}

// ParseAccessToken 은 AT 의 서명·만료를 확인하고 accountID 를 꺼낸다.
func (p *TokenProvider) ParseAccessToken(tokenString string) (int64, error) {
	claims := &jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (any, error) {
			return p.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return 0, err
	}

	accountID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}

	return accountID, nil
}

// NewRefreshToken 은 앱에 줄 원문, DB 에 넣을 해시, 만료 시각을 같이 만든다.
func (p *TokenProvider) NewRefreshToken() (plain string, hash string, expiresAt time.Time) {
	b := make([]byte, 32)
	rand.Read(b)

	plain = base64.RawURLEncoding.EncodeToString(b)
	return plain, p.HashRefreshToken(plain), time.Now().Add(p.refreshTTL)
}

// HashRefreshToken 은 RT 원문을 DB 에 넣을 모양(sha256 → 16진수 글자)으로 바꾼다.
func (p *TokenProvider) HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
