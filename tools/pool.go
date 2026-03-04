package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

// registerPoolTools registers all pool-related MCP tools onto the server.
func registerPoolTools(s *mcp.Server, client *truenas.Client) {
	type listPoolsInput struct {
		Limit  int `json:"limit,omitempty"  jsonschema:"Maximum number of pools to return; 0 means no limit"`
		Offset int `json:"offset,omitempty" jsonschema:"Number of pools to skip; 0 means start from the beginning"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_pools",
		Description: "List all ZFS storage pools and their status, size, and health.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p listPoolsInput) (*mcp.CallToolResult, any, error) {
		pools, err := client.ListPools(ctx, truenas.ListOptions{Limit: p.Limit, Offset: p.Offset})
		if err != nil {
			return nil, nil, fmt.Errorf("list_pools: %w", err)
		}
		return jsonResult(pools)
	})

	type getPoolInput struct {
		ID int `json:"id" jsonschema:"Numeric pool ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_pool",
		Description: "Get detailed information about a specific ZFS pool by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p getPoolInput) (*mcp.CallToolResult, any, error) {
		if p.ID <= 0 {
			return nil, nil, errors.New("get_pool: id must be a positive integer")
		}
		pool, err := client.GetPool(ctx, p.ID)
		if err != nil {
			if errors.Is(err, truenas.ErrNotFound) {
				return nil, nil, fmt.Errorf("get_pool: pool %d not found", p.ID)
			}
			return nil, nil, fmt.Errorf("get_pool: %w", err)
		}
		return jsonResult(pool)
	})
}
