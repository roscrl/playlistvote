package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	refreshWaitTime  = 5 * time.Second
	softExpiryBuffer = 10 * time.Second
)

type token struct {
	Client        *http.Client
	ClientID      string
	ClientSecret  string
	TokenEndpoint string
	Now           func() time.Time

	accessToken   string
	softExpiry    time.Time
	hardExpiry    time.Time
	mutex         sync.Mutex
	done          chan struct{}
	firstInitDone chan struct{}
}

func (t *token) AccessToken() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.accessToken
}

func (t *token) startRefreshLoop(ctx context.Context) {
	firstInit := true

	for {
		select {
		case <-time.After(t.timeToRefresh()):
			err := t.refreshToken(ctx)
			if err != nil {
				log.Printf("refreshing token: %v", err)
				time.Sleep(refreshWaitTime)

				continue
			}

			if firstInit {
				close(t.firstInitDone)

				firstInit = false
			}
		case <-t.done:
			return
		}
	}
}

func (t *token) timeToRefresh() time.Duration {
	now := t.Now()
	if now.After(t.softExpiry) {
		return 0
	}

	return t.softExpiry.Sub(now)
}

func (t *token) stopRefreshLoop() {
	close(t.done)
}

func (t *token) refreshToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.TokenEndpoint, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.ClientID, t.ClientSecret)

	resp, err := t.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	response := struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.accessToken = response.AccessToken

	now := t.Now()

	t.softExpiry = now.Add((time.Duration(response.ExpiresIn) * time.Second) - softExpiryBuffer)
	t.hardExpiry = now.Add(time.Duration(response.ExpiresIn) * time.Second)

	return nil
}
