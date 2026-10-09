package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newTestServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0.0.1"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: "ping", Description: "Returns pong."},
		func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "pong"}}}, nil, nil
		})
	return s
}

func TestHTTPHandlerServesTools(t *testing.T) {
	ts := httptest.NewServer(newHTTPHandler(newTestServer()))
	defer ts.Close()

	ctx := context.Background()
	c := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	cs, err := c.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() {
		if err := cs.Close(); err != nil {
			t.Logf("close: %v", err)
		}
	}()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "ping"})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError || len(res.Content) != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if tc, ok := res.Content[0].(*mcp.TextContent); !ok || tc.Text != "pong" {
		t.Errorf("content = %#v, want pong", res.Content[0])
	}
}

func TestHTTPHandlerRejectsOversizedBody(t *testing.T) {
	ts := httptest.NewServer(newHTTPHandler(newTestServer()))
	defer ts.Close()

	body := bytes.Repeat([]byte("a"), 4<<20+1)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("close body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

func TestRequireEnv(t *testing.T) {
	t.Setenv("TRUENAS_MCP_TEST_VAR", "value")
	if got, err := requireEnv("TRUENAS_MCP_TEST_VAR"); err != nil || got != "value" {
		t.Errorf("requireEnv = %q, %v; want value, nil", got, err)
	}

	t.Setenv("TRUENAS_MCP_TEST_VAR", "")
	if _, err := requireEnv("TRUENAS_MCP_TEST_VAR"); err == nil {
		t.Error("requireEnv with empty value: want error, got nil")
	}
}
