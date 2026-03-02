package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

// registerNetworkTools registers network-related MCP tools onto the server.
func registerNetworkTools(s *mcp.Server, client *truenas.Client) {
	type listInterfacesInput struct{}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_interfaces",
		Description: "List all network interfaces configured on the TrueNAS host, including bridges, physical ports, bonds, and VLANs. Useful for finding bridge names when attaching NICs to VMs.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listInterfacesInput) (*mcp.CallToolResult, any, error) {
		ifaces, err := client.ListInterfaces(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_interfaces: %w", err)
		}
		return jsonResult(ifaces)
	})
}
