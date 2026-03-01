package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

// registerDestructiveTools adds opt-in destructive tools to the server.
// These tools are only registered when TRUENAS_ALLOW_DESTRUCTIVE=true.
func registerDestructiveTools(s *mcp.Server, client *truenas.Client) {
	destructiveHint := true
	type deleteVMInput struct {
		ID        int  `json:"id"        jsonschema:"required,Numeric VM ID"`
		Confirmed bool `json:"confirmed" jsonschema:"required,Must be set to true to confirm deletion"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_vm",
		Description: "Permanently delete a virtual machine by ID. The VM must be stopped first. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
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

	type deleteAppInput struct {
		Name      string `json:"name"      jsonschema:"required,App name (as shown in the TrueNAS UI)"`
		Confirmed bool   `json:"confirmed" jsonschema:"required,Must be set to true to confirm deletion"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_app",
		Description: "Permanently delete a TrueNAS app by name. The app must be stopped first. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p deleteAppInput) (*mcp.CallToolResult, any, error) {
		if !p.Confirmed {
			return nil, nil, errors.New("delete_app: confirmed must be true to proceed with deletion")
		}
		if err := client.DeleteApp(ctx, p.Name); err != nil {
			return nil, nil, fmt.Errorf("delete_app: %w", err)
		}
		return jsonResult(map[string]any{"deleted": true, "name": p.Name})
	})

	type deleteSnapshotInput struct {
		ID        string `json:"id"        jsonschema:"required,Full snapshot ID in dataset@name form, e.g. Storage/backups@before-upgrade"`
		Confirmed bool   `json:"confirmed" jsonschema:"required,Must be set to true to confirm deletion"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_snapshot",
		Description: "Permanently delete a ZFS snapshot. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p deleteSnapshotInput) (*mcp.CallToolResult, any, error) {
		if !p.Confirmed {
			return nil, nil, errors.New("delete_snapshot: confirmed must be true to proceed with deletion")
		}
		if err := client.DeleteSnapshot(ctx, p.ID); err != nil {
			return nil, nil, fmt.Errorf("delete_snapshot: %w", err)
		}
		return jsonResult(map[string]any{"deleted": true, "id": p.ID})
	})
}
