package login

import (
	"errors"
	"fmt"
	"net"
	"net/http"
)

func listenLoopback() (net.Listener, string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", err
	}

	port := ln.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	return ln, redirectURL, nil
}

type callbackResult struct {
	code string
	err  error
}

func serveCallback(ln net.Listener, state string) (<-chan callbackResult, func()) {
	results := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var res callbackResult
		switch {
		case q.Get("error") != "":
			res.err = fmt.Errorf("구글 로그인 취소 또는 실패: %s", q.Get("error"))
		case q.Get("state") != state:
			res.err = errors.New("state 가 달라요")
		default:
			res.code = q.Get("code")
		}

		if res.err != nil {
			fmt.Fprint(w, "로그인에 실패했어요. 아지트로 돌아가 주세요.")
		} else {
			fmt.Fprint(w, "로그인됐어요. 이 창은 닫고 아지트로 돌아가세요.")
		}

		select {
		case results <- res:
		default:
		}
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)

	return results, func() { srv.Close() }
}
