package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a Client pointed at the provided test server URL.
func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	c, err := NewClient(serverURL, "test-api-key", false) //nolint:gosec // G101: test-api-key is a fake placeholder, not a real credential
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestNewClient_validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		apiURL  string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "all valid",
			apiURL:  "https://truenas.local/api/v2.0",
			apiKey:  "test-api-key", //nolint:gosec // G101: fake placeholder
			wantErr: false,
		},
		{
			name:    "empty apiURL",
			apiURL:  "",
			apiKey:  "test-api-key", //nolint:gosec // G101: fake placeholder
			wantErr: true,
		},
		{
			name:    "empty apiKey",
			apiURL:  "https://truenas.local/api/v2.0",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewClient(tc.apiURL, tc.apiKey, false)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestClient_authHeader(t *testing.T) {
	t.Parallel()

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"version": "TrueNAS-SCALE-24.10"}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var out map[string]string
	if err := c.get(context.Background(), "/system/version", &out); err != nil {
		t.Fatalf("get: %v", err)
	}

	const wantAuth = "Bearer test-api-key" //nolint:gosec // G101: fake placeholder compared in test
	if gotAuth != wantAuth {
		t.Errorf("Authorization header = %q, want %q", gotAuth, wantAuth)
	}
}

func TestClient_404_returnsErrNotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var out map[string]any
	err := c.get(context.Background(), "/nonexistent", &out)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestClient_4xx_returnsAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Invalid API key"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var out map[string]any
	err := c.get(context.Background(), "/system/info", &out)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}
