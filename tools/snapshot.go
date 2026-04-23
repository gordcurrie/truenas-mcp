package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerSnapshotTools registers all ZFS snapshot-related MCP tools onto the server.
func registerSnapshotTools(s *mcp.Server, client truenasClient) {
	type listSnapshotsInput struct {
		// Dataset filters results to snapshots of this dataset (e.g. "Storage/backups").
		// Leave empty to return snapshots from all datasets.
		Dataset string `json:"dataset,omitempty" jsonschema:"Full dataset path to filter by; leave empty for all datasets"`
		Limit   int    `json:"limit,omitempty"   jsonschema:"Maximum number of snapshots to return; 0 means no limit"`
		Offset  int    `json:"offset,omitempty"  jsonschema:"Number of snapshots to skip; 0 means start from the beginning"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_snapshots",
		Description: "List ZFS snapshots. Optionally filter by dataset path (e.g. \"Storage/backups\").",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p listSnapshotsInput) (*mcp.CallToolResult, any, error) {
		snaps, err := client.ListSnapshots(ctx, p.Dataset, truenas.ListOptions{Limit: p.Limit, Offset: p.Offset})
		if err != nil {
			return errorResult(fmt.Errorf("list_snapshots: %w", err))
		}
		return jsonResult(snaps)
	})

	type getSnapshotInput struct {
		// ID is the full snapshot identifier in "dataset@name" form,
		// e.g. "Storage/backups@before-upgrade".
		ID string `json:"id" jsonschema:"Full snapshot ID in dataset@name form, e.g. Storage/backups@before-upgrade"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_snapshot",
		Description: "Get detailed information about a specific ZFS snapshot by its full ID (e.g. \"Storage/backups@before-upgrade\").",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p getSnapshotInput) (*mcp.CallToolResult, any, error) {
		if p.ID == "" {
			return errorResult(errors.New("get_snapshot: id must not be empty"))
		}
		snap, err := client.GetSnapshot(ctx, p.ID)
		if err != nil {
			if errors.Is(err, truenas.ErrNotFound) {
				return errorResult(fmt.Errorf("get_snapshot: snapshot %q not found", p.ID))
			}
			return errorResult(fmt.Errorf("get_snapshot: %w", err))
		}
		return jsonResult(snap)
	})

	type createSnapshotInput struct {
		// Dataset is the full dataset path, e.g. "Storage/backups". Required.
		Dataset string `json:"dataset" jsonschema:"Full dataset path including pool, e.g. Storage/backups"`
		// Name is the snapshot suffix (the part after @). Required.
		Name string `json:"name" jsonschema:"Snapshot name (suffix after @), e.g. before-upgrade"`
		// Recursive also snapshots all descendant datasets when true.
		Recursive bool `json:"recursive,omitempty" jsonschema:"Also snapshot all descendant datasets"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_snapshot",
		Description: "Create a new ZFS snapshot of a dataset. Provide the full dataset path and a snapshot name.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p createSnapshotInput) (*mcp.CallToolResult, any, error) {
		if p.Dataset == "" {
			return errorResult(errors.New("create_snapshot: dataset must not be empty"))
		}
		if p.Name == "" {
			return errorResult(errors.New("create_snapshot: name must not be empty"))
		}
		snap, err := client.CreateSnapshot(ctx, truenas.CreateSnapshotParams{
			Dataset:   p.Dataset,
			Name:      p.Name,
			Recursive: p.Recursive,
		})
		if err != nil {
			return errorResult(fmt.Errorf("create_snapshot: %w", err))
		}
		return jsonResult(snap)
	})
}
