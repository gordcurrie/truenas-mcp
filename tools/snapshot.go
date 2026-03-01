package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

// registerSnapshotTools registers all ZFS snapshot-related MCP tools onto the server.
func registerSnapshotTools(s *mcp.Server, client *truenas.Client) {
	type listSnapshotsInput struct {
		// Dataset filters results to snapshots of this dataset (e.g. "Storage/backups").
		// Leave empty to return snapshots from all datasets.
		Dataset string `json:"dataset,omitempty" jsonschema:"Full dataset path to filter by; leave empty for all datasets"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_snapshots",
		Description: "List ZFS snapshots. Optionally filter by dataset path (e.g. \"Storage/backups\").",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p listSnapshotsInput) (*mcp.CallToolResult, any, error) {
		snaps, err := client.ListSnapshots(ctx, p.Dataset)
		if err != nil {
			return nil, nil, fmt.Errorf("list_snapshots: %w", err)
		}
		return jsonResult(snaps)
	})

	type getSnapshotInput struct {
		// ID is the full snapshot identifier in "dataset@name" form,
		// e.g. "Storage/backups@before-upgrade".
		ID string `json:"id" jsonschema:"required,Full snapshot ID in dataset@name form, e.g. Storage/backups@before-upgrade"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_snapshot",
		Description: "Get detailed information about a specific ZFS snapshot by its full ID (e.g. \"Storage/backups@before-upgrade\").",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p getSnapshotInput) (*mcp.CallToolResult, any, error) {
		snap, err := client.GetSnapshot(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_snapshot: %w", err)
		}
		return jsonResult(snap)
	})

	type createSnapshotInput struct {
		// Dataset is the full dataset path, e.g. "Storage/backups". Required.
		Dataset string `json:"dataset" jsonschema:"required,Full dataset path including pool, e.g. Storage/backups"`
		// Name is the snapshot suffix (the part after @). Required.
		Name string `json:"name" jsonschema:"required,Snapshot name (suffix after @), e.g. before-upgrade"`
		// Recursive also snapshots all descendant datasets when true.
		Recursive bool `json:"recursive,omitempty" jsonschema:"Also snapshot all descendant datasets"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_snapshot",
		Description: "Create a new ZFS snapshot of a dataset. Provide the full dataset path and a snapshot name.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p createSnapshotInput) (*mcp.CallToolResult, any, error) {
		snap, err := client.CreateSnapshot(ctx, truenas.CreateSnapshotParams{
			Dataset:   p.Dataset,
			Name:      p.Name,
			Recursive: p.Recursive,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create_snapshot: %w", err)
		}
		return jsonResult(snap)
	})

	type rollbackSnapshotInput struct {
		// ID is the full snapshot identifier, e.g. "Storage/backups@before-upgrade". Required.
		ID string `json:"id" jsonschema:"required,Full snapshot ID in dataset@name form, e.g. Storage/backups@before-upgrade"`
		// Recursive destroys more recent snapshots on the dataset and descendants before rolling back.
		Recursive bool `json:"recursive,omitempty" jsonschema:"Destroy more recent snapshots before rollback"`
		// RecursiveClones also destroys more recent clones, used together with Recursive.
		RecursiveClones bool `json:"recursive_clones,omitempty" jsonschema:"Also destroy more recent clones (requires recursive=true)"`
		// Force rolls back even if the dataset is currently busy.
		Force bool `json:"force,omitempty" jsonschema:"Force rollback even if dataset is busy"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rollback_snapshot",
		Description: "Roll a dataset back to a previous ZFS snapshot. Any changes made after the snapshot will be lost.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p rollbackSnapshotInput) (*mcp.CallToolResult, any, error) {
		if p.RecursiveClones && !p.Recursive {
			return nil, nil, fmt.Errorf("rollback_snapshot: recursive_clones requires recursive=true")
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
