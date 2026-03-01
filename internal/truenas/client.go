package truenas

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

// Client is a TrueNAS SCALE API client that authenticates using an API key.
// All methods require a context.Context as the first parameter.
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewClient creates a new TrueNAS SCALE API client.
//
// apiURL must be the full base URL including the API version prefix,
// e.g. "https://truenas.local/api/v2.0".
// apiKey is the API key created in TrueNAS UI under Credentials → API Keys.
// insecure, when true, disables TLS certificate verification — only use this
// when your TrueNAS host uses a self-signed certificate and you understand the risks.
func NewClient(apiURL, apiKey string, insecure bool) (*Client, error) {
	if apiURL == "" {
		return nil, fmt.Errorf("apiURL must not be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey must not be empty")
	}

	transport := &http.Transport{}
	if insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, /* #nosec G402 */ //nolint:gosec // G402: only set when PROXMOX_INSECURE=true; user must explicitly opt in
		}
	}

	return &Client{
		baseURL: apiURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   defaultTimeout,
		},
		apiKey: apiKey,
	}, nil
}

// get performs an authenticated GET request to path (relative to baseURL)
// and JSON-decodes the result into result (which must be a pointer).
func (c *Client) get(ctx context.Context, path string, result any) error {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing get response body", "err", err)
		}
	}()

	return c.decode(resp.Body, result)
}

// post performs an authenticated POST request to path (relative to baseURL).
// The Proxmox {"data": ...} envelope is unwrapped and decoded into result.
// To POST with a request body, use postWithBody.
func (c *Client) post(ctx context.Context, path string, result any) error {
	resp, err := c.do(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing post response body", "err", err)
		}
	}()

	return c.decode(resp.Body, result)
}

// postWithBody performs an authenticated POST request to path with body
// marshalled as JSON, and decodes the response into result.
func (c *Client) postWithBody(ctx context.Context, path string, body, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshalling request body: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing postWithBody response body", "err", err)
		}
	}()

	return c.decode(resp.Body, result)
}

// delete performs an authenticated DELETE request to path (relative to baseURL).
// The response body is drained and closed; no result decoding is performed.
func (c *Client) delete(ctx context.Context, path string) error {
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing delete response body", "err", err)
		}
	}()

	return nil
}

// put performs an authenticated PUT request to path with body marshalled as
// JSON and decodes the response into result.
func (c *Client) put(ctx context.Context, path string, body, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshalling request body: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPut, path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing put response body", "err", err)
		}
	}()

	return c.decode(resp.Body, result)
}

// do constructs and executes an HTTP request, handling auth headers and error
// status codes. The caller is responsible for closing resp.Body.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("creating request %s %s: %w", method, path, err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req) /* #nosec G704 */ //nolint:gosec // G704: URL is constructed from PROXMOX_API_URL which the user must explicitly supply
	if err != nil {
		return nil, fmt.Errorf("executing request %s %s: %w", method, path, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing not-found response body", "err", err)
		}
		return nil, ErrNotFound
	}

	if resp.StatusCode >= http.StatusBadRequest {
		rawBody, _ := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing error response body", "err", err)
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(rawBody)}
	}

	return resp, nil
}

// decode reads r and JSON-decodes the response directly into result.
// Unlike Proxmox, TrueNAS responses are not wrapped in a {"data": ...} envelope.
func (c *Client) decode(r io.Reader, result any) error {
	if result == nil {
		return nil
	}
	if err := json.NewDecoder(r).Decode(result); err != nil {
		return fmt.Errorf("decoding API response: %w", err)
	}
	return nil
}
