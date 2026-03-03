package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

// registerDatasetTools registers all dataset-related MCP tools onto the server.
func registerDatasetTools(s *mcp.Server, client *truenas.Client) {
	type listDatasetsInput struct {
		// Pool filters the results to datasets belonging to this pool name.
		// Leave empty to return datasets from all pools.
		Pool string `json:"pool,omitempty" jsonschema:"Pool name to filter by; leave empty for all pools"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_datasets",
		Description: "List ZFS datasets and zvols. Optionally filter by pool name.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p listDatasetsInput) (*mcp.CallToolResult, any, error) {
		datasets, err := client.ListDatasets(ctx, p.Pool)
		if err != nil {
			return nil, nil, fmt.Errorf("list_datasets: %w", err)
		}
		return jsonResult(datasets)
	})

	type getDatasetInput struct {
		// ID is the full ZFS path, e.g. "Storage/backups".
		ID string `json:"id" jsonschema:"Full dataset path including pool, e.g. Storage/backups"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_dataset",
		Description: "Get detailed information about a specific ZFS dataset or zvol by its full path ID (e.g. \"Storage/backups\").",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p getDatasetInput) (*mcp.CallToolResult, any, error) {
		if p.ID == "" {
			return nil, nil, errors.New("get_dataset: id must not be empty")
		}
		dataset, err := client.GetDataset(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_dataset: %w", err)
		}
		return jsonResult(dataset)
	})

	type createDatasetInput struct {
		// Name is the full dataset path including the pool, e.g. "Storage/backups".
		Name string `json:"name" jsonschema:"Full dataset path including pool, e.g. Storage/backups"`
		// Type is FILESYSTEM (default) or VOLUME.
		Type string `json:"type,omitempty" jsonschema:"Dataset type: FILESYSTEM or VOLUME; defaults to FILESYSTEM"`
		// Compression algorithm, e.g. lz4, gzip, zstd. Omit to use the pool default.
		Compression string `json:"compression,omitempty" jsonschema:"Compression algorithm, e.g. lz4, gzip, zstd"`
		// Comments is a free-text description for the dataset.
		Comments string `json:"comments,omitempty" jsonschema:"Free-text description for the dataset"`
		// Quota is the maximum space in bytes for this dataset and its children (0 = no quota).
		Quota int64 `json:"quota,omitempty" jsonschema:"Maximum space in bytes (0 = no quota)"`
		// Volsize is the size in bytes of the zvol block device. Required when type is VOLUME.
		Volsize int64 `json:"volsize,omitempty" jsonschema:"Size in bytes of the zvol (required when type is VOLUME)"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_dataset",
		Description: "Create a new ZFS dataset (filesystem) or zvol. At minimum, provide the full path name including the pool (e.g. \"Storage/backups\"). For zvols (type=VOLUME) volsize in bytes is required.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p createDatasetInput) (*mcp.CallToolResult, any, error) {
		if p.Name == "" {
			return nil, nil, errors.New("create_dataset: name must not be empty")
		}
		dataset, err := client.CreateDataset(ctx, &truenas.CreateDatasetParams{
			Name:        p.Name,
			Type:        p.Type,
			Compression: p.Compression,
			Comments:    p.Comments,
			Quota:       p.Quota,
			Volsize:     p.Volsize,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create_dataset: %w", err)
		}
		return jsonResult(dataset)
	})
}
