package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerSystemTools adds system info MCP tools to the server.
func registerSystemTools(s *mcp.Server, client truenasClient) {
	type getSystemInfoInput struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_system_info",
		Description: "Get general system information from the TrueNAS SCALE host, including version, hostname, CPU model, memory, uptime, and load averages.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ getSystemInfoInput) (*mcp.CallToolResult, any, error) {
		info, err := client.GetSystemInfo(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("get_system_info: %w", err))
		}
		return jsonResult(info)
	})
}
