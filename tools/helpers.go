// Package tools registers all MCP tools onto the server.
package tools

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// jsonResult marshals v to indented JSON and returns it as a TextContent result.
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

// errorResult returns an MCP-spec-compliant tool error result (isError: true).
// Per MCP spec §6, tool execution errors should use isError rather than
// propagating a protocol-level error.
func errorResult(err error) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
	}, nil, nil
}
