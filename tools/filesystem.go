package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerFilesystemTools registers filesystem-related MCP tools onto the server.
func registerFilesystemTools(s *mcp.Server, client truenasClient) {
	type listDirectoryInput struct {
		// Path is the absolute path on the TrueNAS host to list, e.g. "/mnt/Storage/pbs".
		Path string `json:"path" jsonschema:"Absolute path on the TrueNAS host filesystem, e.g. /mnt/Storage/pbs"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_directory",
		Description: "List the contents of a directory on the TrueNAS host filesystem. Returns name, type (FILE or DIRECTORY), size, and path for each entry.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p listDirectoryInput) (*mcp.CallToolResult, any, error) {
		if p.Path == "" {
			return errorResult(errors.New("list_directory: path must not be empty"))
		}
		entries, err := client.ListDirectory(ctx, p.Path)
		if err != nil {
			return errorResult(fmt.Errorf("list_directory: %w", err))
		}
		return jsonResult(entries)
	})
}
