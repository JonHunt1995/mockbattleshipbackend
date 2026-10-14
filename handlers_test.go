package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JonHunt1995/mockbattleshipbackend/store"
)

func TestCreateNewGame_InviteURL(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		xOriginalHost  string
		expectedPrefix string
		expectedDomain string
	}{
		{
			name:           "localhost with port",
			host:           "localhost:4000",
			expectedPrefix: "http://",
			expectedDomain: "localhost:4000",
		},
		{
			name:           "localhost without port",
			host:           "localhost",
			expectedPrefix: "http://",
			expectedDomain: "localhost",
		},
		{
			name:           "remote host with port",
			host:           "api.example.com:443",
			expectedPrefix: "https://",
			expectedDomain: "api.example.com",
		},
		{
			name:           "remote host without port",
			host:           "api.example.com",
			expectedPrefix: "https://",
			expectedDomain: "api.example.com",
		},
		{
			name:           "X-Original-Host overrides Host",
			host:           "backend-internal:8080",
			xOriginalHost:  "localhost:5173",
			expectedPrefix: "http://",
			expectedDomain: "localhost:5173",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := &application{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				store:  store.NewMemGameStore(),
			}

			req := httptest.NewRequest(http.MethodPost, "/api/newgame", nil)
			req.Host = tc.host
			if tc.xOriginalHost != "" {
				req.Header.Set("X-Original-Host", tc.xOriginalHost)
			}
			rec := httptest.NewRecorder()

			app.createNewGame(rec, req)

			if rec.Code != http.StatusAccepted {
				t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
			}

			var resp createGameResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			expectedPrefix := tc.expectedPrefix + tc.expectedDomain + "/setup/"
			if !strings.HasPrefix(resp.InviteLink, expectedPrefix) {
				t.Errorf("expected link to start with %q, got %q", expectedPrefix, resp.InviteLink)
			}
		})
	}
}
