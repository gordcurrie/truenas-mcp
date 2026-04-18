package truenas

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const wsHandshakeTimeout = 10 * time.Second

// rpcRequest is a JSON-RPC 2.0 request frame sent to TrueNAS.
type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// rpcResponse is a JSON-RPC 2.0 response frame received from TrueNAS.
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

// rpcError is the error object in a JSON-RPC 2.0 error response.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ErrName string `json:"errname"`
		Reason  string `json:"reason"`
	} `json:"data"`
}

// Client is a TrueNAS SCALE API client that communicates via WebSocket JSON-RPC 2.0.
// All methods require a context.Context as the first parameter.
// Call Connect before using any API methods.
type Client struct {
	host      string // base URL, e.g. "https://truenas.local"
	apiKey    string
	insecure  bool
	conn      *websocket.Conn
	mu        sync.Mutex // protects conn writes and pending map
	nextID    atomic.Int64
	pending   map[int64]chan rpcResponse
	done      chan struct{}
	closeOnce sync.Once
}

// NewClient creates a new TrueNAS SCALE API client.
//
// host is the base URL of the TrueNAS host, e.g. "https://truenas.local".
// The client derives the WebSocket endpoint (wss://truenas.local/api/current) automatically.
// insecure, when true, disables TLS certificate verification — only use this when your
// TrueNAS host uses a self-signed certificate and you understand the risks.
// Call Connect(ctx) before using any API methods.
func NewClient(host, apiKey string, insecure bool) (*Client, error) {
	if host == "" {
		return nil, fmt.Errorf("host must not be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey must not be empty")
	}
	return &Client{
		host:     strings.TrimRight(host, "/"),
		apiKey:   apiKey,
		insecure: insecure,
		pending:  make(map[int64]chan rpcResponse),
		done:     make(chan struct{}),
	}, nil
}

// wsURL derives the WebSocket URL from the HTTP base URL.
//
//	https://nas.example.com → wss://nas.example.com/api/current
//	http://nas.example.com  → ws://nas.example.com/api/current
func (c *Client) wsURL() string {
	u := strings.TrimRight(c.host, "/")
	switch {
	case strings.HasPrefix(u, "https://"):
		return "wss://" + strings.TrimPrefix(u, "https://") + "/api/current"
	case strings.HasPrefix(u, "http://"):
		return "ws://" + strings.TrimPrefix(u, "http://") + "/api/current"
	default:
		return "wss://" + u + "/api/current"
	}
}

// Connect opens the WebSocket connection to TrueNAS and authenticates with the API key.
// Must be called before any API methods. Returns an error if the connection or
// authentication fails.
func (c *Client) Connect(ctx context.Context) error {
	wsAddr := c.wsURL()

	dialer := websocket.Dialer{
		HandshakeTimeout: wsHandshakeTimeout,
	}
	if c.insecure {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} /* #nosec G402 */ //nolint:gosec // G402: InsecureSkipVerify is only set when TRUENAS_INSECURE=true, which the user must explicitly opt into. Default is secure (verify enabled).
	}

	conn, resp, err := dialer.DialContext(ctx, wsAddr, nil) //nolint:bodyclose // WebSocket upgrade response body is managed by conn after successful upgrade
	if err != nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return fmt.Errorf("connecting to TrueNAS at %s: %w", wsAddr, err)
	}
	c.conn = conn

	go c.readLoop()

	// Authenticate with the API key.
	var ok bool
	if err := c.call(ctx, "auth.login_with_api_key", []any{c.apiKey}, &ok); err != nil {
		_ = c.Close() // best-effort cleanup; auth error takes precedence
		return fmt.Errorf("authenticating with TrueNAS: %w", err)
	}
	if !ok {
		_ = c.Close() // best-effort cleanup
		return fmt.Errorf("TrueNAS rejected the API key")
	}
	return nil
}

// Close closes the WebSocket connection and unblocks any pending calls.
func (c *Client) Close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		close(c.done)
		if c.conn != nil {
			closeErr = c.conn.Close()
		}
	})
	return closeErr
}

// call sends a JSON-RPC 2.0 request and decodes the result into out.
// out may be nil if the caller does not need the result.
func (c *Client) call(ctx context.Context, method string, params, out any) error {
	id := c.nextID.Add(1)

	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshalling RPC request for %s: %w", method, err)
	}

	ch := make(chan rpcResponse, 1)

	c.mu.Lock()
	c.pending[id] = ch
	err = c.conn.WriteMessage(websocket.TextMessage, data)
	c.mu.Unlock()

	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return fmt.Errorf("sending RPC request for %s: %w", method, err)
	}

	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return fmt.Errorf("RPC call %s: %w", method, ctx.Err())
	case <-c.done:
		return fmt.Errorf("RPC call %s: connection closed", method)
	case resp := <-ch:
		if resp.Error != nil {
			return c.mapRPCError(method, resp.Error)
		}
		if out == nil {
			return nil
		}
		if err := json.Unmarshal(resp.Result, out); err != nil {
			return fmt.Errorf("decoding RPC result for %s: %w", method, err)
		}
		return nil
	}
}

// readLoop reads WebSocket frames and routes responses to pending callers.
// Runs as a goroutine until the connection is closed or an error occurs.
func (c *Client) readLoop() {
	defer func() {
		c.closeOnce.Do(func() { close(c.done) })
		// Unblock all pending callers with a connection-closed error.
		connClosed := &rpcError{Code: -32000, Message: "connection closed"}
		c.mu.Lock()
		for id, ch := range c.pending {
			ch <- rpcResponse{Error: connClosed}
			delete(c.pending, id)
		}
		c.mu.Unlock()
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var resp rpcResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			continue // skip malformed frames
		}
		if resp.ID == nil {
			continue // server notification — no pending caller
		}
		c.mu.Lock()
		ch, ok := c.pending[*resp.ID]
		if ok {
			delete(c.pending, *resp.ID)
		}
		c.mu.Unlock()
		if ok {
			ch <- resp
		}
	}
}

// mapRPCError converts a JSON-RPC error response into ErrNotFound or *APIError.
func (c *Client) mapRPCError(method string, e *rpcError) error {
	haystack := strings.ToLower(e.Message + " " + e.Data.ErrName + " " + e.Data.Reason)
	if strings.Contains(haystack, "not found") ||
		strings.Contains(haystack, "does not exist") ||
		strings.EqualFold(e.Data.ErrName, "NOT_FOUND") {
		return ErrNotFound
	}
	body := e.Data.Reason
	if body == "" {
		body = e.Message
	}
	return &APIError{StatusCode: e.Code, Body: fmt.Sprintf("%s: %s", method, body)}
}
