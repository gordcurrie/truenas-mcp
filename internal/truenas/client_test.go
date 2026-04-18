package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// wsTestServer creates an httptest.Server that speaks WebSocket JSON-RPC 2.0.
// auth.login_with_api_key is handled automatically (returns true).
// All other methods are dispatched to handlers; unregistered methods return a
// -32601 "method not found" error.
type methodHandler func(params json.RawMessage) (any, *rpcError)

func wsTestServer(t *testing.T, handlers map[string]methodHandler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("WS upgrade: %v", err)
			return
		}
		defer conn.Close() //nolint:errcheck // test cleanup

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return // client disconnected
			}

			// Parse the bare fields we need.
			var frame struct {
				ID     int64           `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			if err := json.Unmarshal(data, &frame); err != nil {
				continue
			}

			var result json.RawMessage
			var rpcErr *rpcError

			if frame.Method == "auth.login_with_api_key" {
				result = json.RawMessage("true")
			} else if h, ok := handlers[frame.Method]; ok {
				v, e := h(frame.Params)
				if e != nil {
					rpcErr = e
				} else {
					result, _ = json.Marshal(v)
				}
			} else {
				rpcErr = &rpcError{Code: -32601, Message: "method not found: " + frame.Method}
			}

			resp := map[string]any{"jsonrpc": "2.0", "id": frame.ID}
			if rpcErr != nil {
				resp["error"] = rpcErr
			} else {
				resp["result"] = result
			}

			out, _ := json.Marshal(resp)
			_ = conn.WriteMessage(websocket.TextMessage, out)
		}
	}))
	return srv
}

// newTestClient creates a Client connected to the provided test server.
func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	c, err := NewClient(serverURL, "test-api-key", false) //nolint:gosec // G101: test-api-key is a fake placeholder, not a real credential
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestNewClient_validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		host    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "all valid",
			host:    "https://truenas.local",
			apiKey:  "test-api-key", //nolint:gosec // G101: fake placeholder
			wantErr: false,
		},
		{
			name:    "empty host",
			host:    "",
			apiKey:  "test-api-key", //nolint:gosec // G101: fake placeholder
			wantErr: true,
		},
		{
			name:    "empty apiKey",
			host:    "https://truenas.local",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "no scheme",
			host:    "truenas.local",
			apiKey:  "test-api-key", //nolint:gosec // G101: fake placeholder
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			host:    "ftp://truenas.local",
			apiKey:  "test-api-key", //nolint:gosec // G101: fake placeholder
			wantErr: true,
		},
		{
			name:    "legacy path stripped",
			host:    "https://truenas.local/api/v2.0",
			apiKey:  "test-api-key", //nolint:gosec // G101: fake placeholder
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewClient(tc.host, tc.apiKey, false)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestClient_wsURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		host string
		want string
	}{
		{"https://truenas.local", "wss://truenas.local/api/current"},
		{"https://truenas.local/", "wss://truenas.local/api/current"},
		{"http://truenas.local", "ws://truenas.local/api/current"},
		{"truenas.local", "wss://truenas.local/api/current"},
	}

	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			t.Parallel()
			c := &Client{host: strings.TrimRight(tc.host, "/")}
			if got := c.wsURL(); got != tc.want {
				t.Errorf("wsURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClient_Connect_rejectsInvalidKey(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close() //nolint:errcheck // test cleanup

		// First message is core.set_options — acknowledge it so Connect proceeds.
		_, data, _ := conn.ReadMessage()
		var frame struct {
			ID     int64  `json:"id"`
			Method string `json:"method"`
		}
		_ = json.Unmarshal(data, &frame)
		if frame.Method == "core.set_options" {
			ack, _ := json.Marshal(map[string]any{
				"jsonrpc": "2.0",
				"id":      frame.ID,
				"result":  nil,
			})
			_ = conn.WriteMessage(websocket.TextMessage, ack)
			// Read the actual auth message next.
			_, data, _ = conn.ReadMessage()
			_ = json.Unmarshal(data, &frame)
		}

		// Reject the auth call.
		resp, _ := json.Marshal(map[string]any{
			"jsonrpc": "2.0",
			"id":      frame.ID,
			"result":  false, // auth rejected
		})
		_ = conn.WriteMessage(websocket.TextMessage, resp)
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "bad-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	err = c.Connect(context.Background())
	if err == nil {
		_ = c.Close()
		t.Fatal("expected Connect to fail for rejected key, got nil")
	}
	if !strings.Contains(err.Error(), "rejected") {
		t.Errorf("error %q does not mention 'rejected'", err)
	}
}

func TestClient_call_success(t *testing.T) {
	t.Parallel()

	want := map[string]string{"version": "TrueNAS-SCALE-26.0"}
	srv := wsTestServer(t, map[string]methodHandler{
		"system.version": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var got map[string]string
	if err := c.call(context.Background(), "system.version", []any{}, &got); err != nil {
		t.Fatalf("call: %v", err)
	}
	if got["version"] != want["version"] {
		t.Errorf("version = %q, want %q", got["version"], want["version"])
	}
}

func TestClient_call_notFound(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.get_instance": func(_ json.RawMessage) (any, *rpcError) {
			return nil, &rpcError{
				Code:    -32001,
				Message: "not found",
			}
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var out map[string]any
	err := c.call(context.Background(), "pool.get_instance", []any{999}, &out)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestClient_call_apiError(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"vm.start": func(_ json.RawMessage) (any, *rpcError) {
			e := &rpcError{Code: -32001, Message: "VM is locked"}
			e.Data.Reason = "VM is locked by another process"
			return nil, e
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var out int
	err := c.call(context.Background(), "vm.start", []any{1}, &out)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != -32001 {
		t.Errorf("StatusCode = %d, want -32001", apiErr.StatusCode)
	}
}

func TestClient_call_contextCancelled(t *testing.T) {
	t.Parallel()

	// Server that never responds after auth.
	srv := wsTestServer(t, map[string]methodHandler{})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := c.call(ctx, "system.info", []any{}, nil)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled in error chain, got %v", err)
	}
}
