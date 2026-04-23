package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestJsonResult_success(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	result, extra, err := jsonResult(payload{Name: "truenas", Age: 3})
	if err != nil {
		t.Fatalf("jsonResult returned error: %v", err)
	}
	if extra != nil {
		t.Errorf("expected nil extra, got %v", extra)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}

	var got payload
	if err := json.Unmarshal([]byte(tc.Text), &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if got.Name != "truenas" || got.Age != 3 {
		t.Errorf("got %+v, want {Name:truenas Age:3}", got)
	}
	// Verify compact formatting — Marshal produces no newlines.
	if strings.Contains(tc.Text, "\n") {
		t.Errorf("expected compact JSON output (no newlines), got: %s", tc.Text)
	}
}

func TestJsonResult_marshalError(t *testing.T) {
	t.Parallel()

	// Channels cannot be marshalled to JSON — exercises the error path.
	ch := make(chan int)
	_, _, err := jsonResult(ch)
	if err == nil {
		t.Fatal("expected marshal error for channel type, got nil")
	}
}

func TestErrorResult(t *testing.T) {
	t.Parallel()

	result, extra, err := errorResult(fmt.Errorf("something went wrong: %w", fmt.Errorf("root cause")))
	if err != nil {
		t.Fatalf("errorResult returned protocol error: %v", err)
	}
	if extra != nil {
		t.Errorf("expected nil extra, got %v", extra)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(tc.Text, "something went wrong") {
		t.Errorf("error text %q does not contain expected message", tc.Text)
	}
}
