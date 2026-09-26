package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const CallbackAddr = "127.0.0.1:1061"

type AuthorizationCallback struct {
	Code  string
	State string
}

func WaitForAuthorizationCallback(ctx context.Context, expectedState string) (AuthorizationCallback, error) {
	listener, err := net.Listen("tcp", CallbackAddr)
	if err != nil {
		return AuthorizationCallback{}, err
	}
	defer listener.Close()

	return waitForAuthorizationCallback(ctx, listener, expectedState)
}

func waitForAuthorizationCallback(ctx context.Context, listener net.Listener, expectedState string) (AuthorizationCallback, error) {
	resCh := make(chan AuthorizationCallback, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		res, err := ParseAuthorizationCallback(r, expectedState)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			errCh <- err
			return
		}

		fmt.Fprintln(w, "Authorization complete. You can close this window.")
		resCh <- res
	})

	serverErrCh := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
	}()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	select {
	case res := <-resCh:
		return res, nil
	case err := <-errCh:
		return AuthorizationCallback{}, err
	case err := <-serverErrCh:
		return AuthorizationCallback{}, err
	case <-ctx.Done():
		return AuthorizationCallback{}, ctx.Err()
	}
}

func ParseAuthorizationCallback(r *http.Request, expectedState string) (AuthorizationCallback, error) {
	expectedState = strings.TrimSpace(expectedState)
	if expectedState == "" {
		return AuthorizationCallback{}, fmt.Errorf("expected state is required")
	}

	query := r.URL.Query()

	if oauthErr := strings.TrimSpace(query.Get("error")); oauthErr != "" {
		return AuthorizationCallback{}, fmt.Errorf("authorization failed: %s", oauthErr)
	}

	code := strings.TrimSpace(query.Get("code"))
	if code == "" {
		return AuthorizationCallback{}, fmt.Errorf("code is required")
	}

	state := strings.TrimSpace(query.Get("state"))
	if state == "" {
		return AuthorizationCallback{}, fmt.Errorf("state is required")
	}
	if state != expectedState {
		return AuthorizationCallback{}, fmt.Errorf("state mismatch")
	}

	return AuthorizationCallback{
		Code:  code,
		State: state,
	}, nil
}
