package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

const (
	googleIssuer   = "https://accounts.google.com"
	googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"
)

type GoogleClient struct {
	cfg      *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewGoogleClient(clientID, clientSecret string) *GoogleClient {
	keySet := oidc.NewRemoteKeySet(context.Background(), googleCertsURL)
	return &GoogleClient{
		verifier: oidc.NewVerifier(googleIssuer, keySet, &oidc.Config{ClientID: clientID}),
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     endpoints.Google,
			Scopes:       []string{"openid", "email", "profile"},
		},
	}
}

var ErrGoogleExchange = errors.New("구글 로그인을 확인하지 못했어요")

func (g *GoogleClient) Exchange(ctx context.Context, code, verifier, redirectURI string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	token, err := g.cfg.Exchange(ctx, code,
		oauth2.VerifierOption(verifier),
		oauth2.SetAuthURLParam("redirect_uri", redirectURI),
	)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGoogleExchange, err)
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok || idToken == "" {
		return "", fmt.Errorf("%w: id_token 없음", ErrGoogleExchange)
	}
	return idToken, nil
}

// GoogleUser 는 검증된 id_token 에서 꺼낸 사용자 정보.
type GoogleUser struct {
	Subject string
	Email   string
	Name    string
}

var ErrGoogleIDToken = errors.New("구글 로그인 정보가 올바르지 않아요")

// Verify 는 id_token 의 서명·발급자·대상·만료를 확인하고 사용자 정보를 꺼낸다.
func (g *GoogleClient) Verify(ctx context.Context, rawIDToken string) (GoogleUser, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	idToken, err := g.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return GoogleUser{}, fmt.Errorf("%w: %v", ErrGoogleIDToken, err)
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return GoogleUser{}, fmt.Errorf("%w: %v", ErrGoogleIDToken, err)
	}
	if !claims.EmailVerified {
		return GoogleUser{}, fmt.Errorf("%w: 이메일 미인증", ErrGoogleIDToken)
	}

	return GoogleUser{Subject: idToken.Subject, Email: claims.Email, Name: claims.Name}, nil
}
