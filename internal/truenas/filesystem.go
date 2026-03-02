package truenas

import (
	"context"
	"fmt"
	"strings"
)

// DirEntry represents a single file or directory returned by ListDirectory.
type DirEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	RealPath string `json:"realpath"`
	// Type is "FILE" or "DIRECTORY".
	Type string `json:"type"`
	Size int64  `json:"size"`
	Mode int    `json:"mode"`
}

// listDirRequest is the request body for POST /filesystem/listdir.
type listDirRequest struct {
	Path string `json:"path"`
}

// ListDirectory returns the contents of the given absolute path on the TrueNAS
// host filesystem (e.g. "/mnt/Storage/pbs").
func (c *Client) ListDirectory(ctx context.Context, path string) ([]DirEntry, error) {
	if path == "" {
		return nil, fmt.Errorf("listing directory: path must not be empty")
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("listing directory: path must be absolute (start with /), got %q", path)
	}
	var entries []DirEntry
	if err := c.postWithBody(ctx, "/filesystem/listdir", listDirRequest{Path: path}, &entries); err != nil {
		return nil, fmt.Errorf("listing directory %q: %w", path, err)
	}
	return entries, nil
}
