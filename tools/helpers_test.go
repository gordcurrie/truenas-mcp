package tools

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestJsonResult_validInput(t *testing.T) {
	t.Parallel()

	result, _, err := jsonResult(map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("jsonResult: %v", err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
}

func TestJsonResult_containsJSON(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	result, _, err := jsonResult(payload{Name: "truenas", Age: 3})
	if err != nil {
		t.Fatalf("jsonResult: %v", err)
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(text.Text, "truenas") {
		t.Errorf("output does not contain expected value: %s", text.Text)
	}
}
