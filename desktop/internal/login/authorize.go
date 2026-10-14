package login

import (
	"context"
	"errors"
	"time"
)

type Grant struct {
	Code        string
	Verifier    string
	RedirectURI string
}

func Authorize(ctx context.Context, clientId string, openBrowser func(string) error) (Grant, error) {
	verifier, challenge, err := newPKCE()
	if err != nil {
		return Grant{}, err
	}

	state, err := randomString(16)
	if err != nil {
		return Grant{}, err
	}

	ln, redirectURI, err := listenLoopback()
	if err != nil {
		return Grant{}, err
	}
	results, stop := serveCallback(ln, state)
	defer stop()

	if err = openBrowser(authURL(clientId, redirectURI, challenge, state)); err != nil {
		return Grant{}, err
	}

	select {
	case res := <-results:
		if res.err != nil {
			return Grant{}, res.err
		}
		return Grant{Code: res.code, Verifier: verifier, RedirectURI: redirectURI}, nil
	case <-ctx.Done():
		return Grant{}, ctx.Err()
	case <-time.After(3 * time.Minute):
		return Grant{}, errors.New("로그인 시간이 지났어요. 다시 시도해 주세요")
	}
}
