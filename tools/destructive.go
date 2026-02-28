package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

// destructiveHint is the *bool value used to advertise DestructiveHint on tools.
var destructiveHint = true

// registerDestructiveTools adds opt-in destructive tools to the server.
// These tools are only registered when TRUENAS_ALLOW_DESTRUCTIVE=true.
func registerDestructiveTools(s *mcp.Server, client *truenas.Client) {
	type deleteVMInput struct {
		ID        int  `json:"id"        jsonschema:"required,Numeric VM ID"`
		Confirmed bool `json:"confirmed" jsonschema:"required,Must be set to true to confirm deletion"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_vm",
		Description: "Permanently delete a virtual machine by ID. The VM must be stopped first. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p deleteVMInput) (*mcp.CallToolResult, any, error) {
		if !p.Confirmed {
			return nil, nil, errors.New("delete_vm: confirmed must be true to proceed with deletion")
		}
		if err := client.DeleteVM(ctx, p.ID); err != nil {
			return nil, nil, fmt.Errorf("delete_vm: %w", err)
		}
		return jsonResult(map[string]any{"deleted": true, "id": p.ID})
	})
}
