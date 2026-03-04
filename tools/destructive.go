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
		ID        int  `json:"id"        jsonschema:"Numeric VM ID"`
		Confirmed bool `json:"confirmed" jsonschema:"Must be set to true to confirm deletion"`
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
		if p.ID <= 0 {
			return nil, nil, errors.New("delete_vm: id must be a positive integer")
		}
		if err := client.DeleteVM(ctx, p.ID); err != nil {
			return nil, nil, fmt.Errorf("delete_vm: %w", err)
		}
		return jsonResult(map[string]any{"deleted": true, "id": p.ID})
	})

	type deleteAppInput struct {
		Name      string `json:"name"      jsonschema:"App name (as shown in the TrueNAS UI)"`
		Confirmed bool   `json:"confirmed" jsonschema:"Must be set to true to confirm deletion"`
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
		if p.Name == "" {
			return nil, nil, errors.New("delete_app: name must not be empty")
		}
		if err := client.DeleteApp(ctx, p.Name); err != nil {
			return nil, nil, fmt.Errorf("delete_app: %w", err)
		}
		return jsonResult(map[string]any{"deleted": true, "name": p.Name})
	})

	type deleteSnapshotInput struct {
		ID        string `json:"id"        jsonschema:"Full snapshot ID in dataset@name form, e.g. Storage/backups@before-upgrade"`
		Confirmed bool   `json:"confirmed" jsonschema:"Must be set to true to confirm deletion"`
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
		if p.ID == "" {
			return nil, nil, errors.New("delete_snapshot: id must not be empty")
		}
		if err := client.DeleteSnapshot(ctx, p.ID); err != nil {
			return nil, nil, fmt.Errorf("delete_snapshot: %w", err)
		}
		return jsonResult(map[string]any{"deleted": true, "id": p.ID})
	})

	type deleteVMDeviceInput struct {
		ID        int  `json:"id"        jsonschema:"Numeric device ID (from list_vm_devices)"`
		Confirmed bool `json:"confirmed" jsonschema:"Must be set to true to confirm removal"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_vm_device",
		Description: "Remove a hardware device from a VM by its device ID. Use list_vm_devices to find device IDs. Changes take effect on next boot. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p deleteVMDeviceInput) (*mcp.CallToolResult, any, error) {
		if !p.Confirmed {
			return nil, nil, errors.New("delete_vm_device: confirmed must be true to proceed with removal")
		}
		if p.ID <= 0 {
			return nil, nil, errors.New("delete_vm_device: id must be a positive integer")
		}
		if err := client.DeleteVMDevice(ctx, p.ID); err != nil {
			return nil, nil, fmt.Errorf("delete_vm_device: %w", err)
		}
		return jsonResult(map[string]any{"deleted": true, "device_id": p.ID})
	})

	type rollbackSnapshotInput struct {
		// ID is the full snapshot identifier, e.g. "Storage/backups@before-upgrade". Required.
		ID string `json:"id" jsonschema:"Full snapshot ID in dataset@name form, e.g. Storage/backups@before-upgrade"`
		// Confirmed must be true to proceed; protects against accidental data loss.
		Confirmed bool `json:"confirmed" jsonschema:"Must be set to true to confirm rollback and acknowledge that all data written after this snapshot will be permanently lost"`
		// Recursive destroys more recent snapshots on the dataset and descendants before rolling back.
		Recursive bool `json:"recursive,omitempty" jsonschema:"Destroy more recent snapshots before rollback"`
		// RecursiveClones also destroys more recent clones, used together with Recursive.
		RecursiveClones bool `json:"recursive_clones,omitempty" jsonschema:"Also destroy more recent clones (requires recursive=true)"`
		// Force rolls back even if the dataset is currently busy.
		Force bool `json:"force,omitempty" jsonschema:"Force rollback even if dataset is busy"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rollback_snapshot",
		Description: "Roll a dataset back to a previous ZFS snapshot. ALL data written after the snapshot will be permanently destroyed. Set confirmed=true to proceed. Only available when TRUENAS_ALLOW_DESTRUCTIVE=true.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p rollbackSnapshotInput) (*mcp.CallToolResult, any, error) {
		if !p.Confirmed {
			return nil, nil, errors.New("rollback_snapshot: confirmed must be true to proceed with rollback")
		}
		if p.ID == "" {
			return nil, nil, errors.New("rollback_snapshot: id must not be empty")
		}
		if p.RecursiveClones && !p.Recursive {
			return nil, nil, errors.New("rollback_snapshot: recursive_clones requires recursive=true")
		}
		err := client.RollbackSnapshot(ctx, p.ID, truenas.RollbackSnapshotParams{
			Recursive:       p.Recursive,
			RecursiveClones: p.RecursiveClones,
			Force:           p.Force,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("rollback_snapshot: %w", err)
		}
		return jsonResult(map[string]any{"rolled_back": true, "id": p.ID})
	})
}
