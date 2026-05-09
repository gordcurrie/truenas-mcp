package truenas

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsHandshakeTimeout = 10 * time.Second
	defaultCallTimeout = 30 * time.Second // applied when ctx has no deadline
	maxMessageBytes    = 10 * 1024 * 1024 // 10 MiB cap on inbound WebSocket messages
)

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
	host     string // base URL, e.g. "https://truenas.local"
	apiKey   string
	insecure bool
	conn     *websocket.Conn
	mu       sync.Mutex // protects conn, done, and pending map
	reconMu  sync.Mutex // serialises reconnection attempts
	nextID   atomic.Int64
	pending  map[int64]chan rpcResponse
	done     chan struct{}
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
	// Normalize the host URL: strip any path so that legacy values like
	// "https://nas.local/api/v2.0" don't produce a double-path WebSocket URL.
	u, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("invalid host URL %q: %w", host, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid host URL %q: must begin with http:// or https://", host)
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	host = u.String()
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
	u, err := url.Parse(strings.TrimRight(c.host, "/"))
	if err != nil {
		return ""
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	default:
		return ""
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/api/current"
	return u.String()
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

	conn, resp, err := dialer.DialContext(ctx, wsAddr, nil)
	if resp != nil {
		_ = resp.Body.Close() // drain upgrade response body; no-op after a successful upgrade
	}
	if err != nil {
		return fmt.Errorf("connecting to TrueNAS at %s: %w", wsAddr, err)
	}

	// Assign the connection under mu and guard against a double-Connect.
	c.mu.Lock()
	if c.conn != nil {
		c.mu.Unlock()
		_ = conn.Close()
		return fmt.Errorf("Connect called on an already-connected client")
	}
	conn.SetReadLimit(maxMessageBytes)
	c.conn = conn
	c.mu.Unlock()

	go c.readLoop()

	// Initialise the session. The official TrueNAS client always calls
	// core.set_options immediately after connecting to configure session
	// behaviour (e.g. new-style job IDs). Skipping it can cause the server to
	// reject subsequent calls on some TrueNAS 25.04+ builds.
	var setOptsResult any
	if err := c.callOnce(ctx, "core.set_options", []any{map[string]any{
		"legacy_jobs": false,
	}}, &setOptsResult); err != nil {
		// Suppress only JSON-RPC -32601 (method not found): older TrueNAS builds
		// don't implement core.set_options and that is safe to ignore. Any other
		// error (connection failure, auth problem, etc.) is a real failure.
		if !errors.Is(err, ErrMethodNotFound) {
			_ = c.Close()
			return fmt.Errorf("initialising TrueNAS session: %w", err)
		}
	}

	// Authenticate with the API key.
	var ok bool
	if err := c.callOnce(ctx, "auth.login_with_api_key", []any{c.apiKey}, &ok); err != nil {
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
// Safe to call multiple times.
func (c *Client) Close() error {
	c.mu.Lock()
	select {
	case <-c.done:
		c.mu.Unlock()
		return nil // already closed
	default:
		close(c.done)
	}
	conn := c.conn
	c.mu.Unlock()
	if conn != nil {
		return conn.Close()
	}
	return nil
}

// reconnect tears down the current (dead) connection and establishes a fresh one.
// It is safe to call concurrently; only one reconnection attempt runs at a time.
func (c *Client) reconnect(ctx context.Context) error {
	c.reconMu.Lock()
	defer c.reconMu.Unlock()

	// Check under mu whether done is still closed; if not, the connection is healthy.
	c.mu.Lock()
	select {
	case <-c.done:
		// connection is closed — proceed with reconnect
	default:
		c.mu.Unlock()
		return nil // already reconnected by a concurrent call
	}
	oldConn := c.conn
	c.conn = nil // cleared so Connect can assign the new connection under mu
	c.pending = make(map[int64]chan rpcResponse)
	c.done = make(chan struct{})
	c.mu.Unlock()

	// Close old connection outside mu to avoid holding the lock during I/O.
	if oldConn != nil {
		_ = oldConn.Close()
	}

	return c.Connect(ctx)
}

// call sends a JSON-RPC 2.0 request and decodes the result into out.
// out may be nil if the caller does not need the result.
// On a broken connection it attempts one transparent reconnect before returning an error.
func (c *Client) call(ctx context.Context, method string, params, out any) error {
	err := c.callOnce(ctx, method, params, out)
	if err == nil {
		return nil
	}
	// If the connection was dropped, attempt one reconnect and retry.
	select {
	case <-c.done:
		if reconnErr := c.reconnect(ctx); reconnErr != nil {
			return fmt.Errorf("sending RPC request for %s: reconnect failed: %w", method, reconnErr)
		}
		return c.callOnce(ctx, method, params, out)
	default:
		return err
	}
}

// callOnce is the core send-and-wait logic used by call.
func (c *Client) callOnce(ctx context.Context, method string, params, out any) error {
	// Ensure a per-call deadline so a silent server or stalled network never
	// blocks the caller indefinitely. A tighter deadline from the caller wins.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultCallTimeout)
		defer cancel()
	}

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
	if c.conn == nil {
		c.mu.Unlock()
		return fmt.Errorf("sending RPC request for %s: client not connected; call Connect first", method)
	}
	// Apply a write deadline derived from the caller's context so WriteMessage
	// doesn't block indefinitely under network backpressure.
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetWriteDeadline(deadline)
	}
	c.pending[id] = ch
	err = c.conn.WriteMessage(websocket.TextMessage, data)
	_ = c.conn.SetWriteDeadline(time.Time{}) // clear deadline regardless of outcome
	c.mu.Unlock()

	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		// Proactively close done so call() triggers a reconnect attempt even
		// if readLoop has not yet observed the broken connection.
		select {
		case <-c.done:
		default:
			close(c.done)
		}
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
	// Capture both the connection and the done channel for this goroutine's lifetime.
	// Using a local conn ensures a stale readLoop reads from its own (closed)
	// connection and cannot interfere with a freshly-reconnected one.
	c.mu.Lock()
	doneCh := c.done
	conn := c.conn
	c.mu.Unlock()

	defer func() {
		connClosed := &rpcError{Code: -32000, Message: "connection closed"}
		c.mu.Lock()
		// Only close and drain if this is still the active readLoop. If
		// reconnect() has replaced c.done, the pending map was already cleared
		// and the new readLoop owns the new done channel.
		if c.done == doneCh {
			select {
			case <-doneCh:
			default:
				close(doneCh)
			}
			for id, ch := range c.pending {
				ch <- rpcResponse{Error: connClosed}
				delete(c.pending, id)
			}
		}
		c.mu.Unlock()
	}()

	for {
		_, data, err := conn.ReadMessage()
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

// mapRPCError converts a JSON-RPC error response into ErrNotFound,
// ErrMethodNotFound, or *APIError.
// JSON-RPC -32601 (method not found) is returned as an error wrapping
// ErrMethodNotFound so callers can distinguish a missing RPC method (client
// bug or older TrueNAS build) from a missing TrueNAS resource.
func (c *Client) mapRPCError(method string, e *rpcError) error {
	body := e.Data.Reason
	if body == "" {
		body = e.Message
	}
	// -32601 is JSON-RPC "method not found" — a protocol-level error, not a
	// missing TrueNAS resource. Wrap ErrMethodNotFound before the substring scan
	// so the "not found" text in its message never produces a false ErrNotFound.
	if e.Code == -32601 {
		return fmt.Errorf("%s: %s: %w", method, body, ErrMethodNotFound)
	}
	haystack := strings.ToLower(e.Message + " " + e.Data.ErrName + " " + e.Data.Reason)
	if strings.Contains(haystack, "not found") ||
		strings.Contains(haystack, "does not exist") ||
		strings.EqualFold(e.Data.ErrName, "NOT_FOUND") {
		return ErrNotFound
	}
	return &APIError{StatusCode: e.Code, Body: fmt.Sprintf("%s: %s", method, body)}
}
