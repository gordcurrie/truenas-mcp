package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

func TestListDirectory(t *testing.T) {
	t.Run("returns entries as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			listDirectoryFn: func(_ context.Context, _ string) ([]truenas.DirEntry, error) {
				return []truenas.DirEntry{{Name: "pbs", Type: "DIRECTORY"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_directory", map[string]any{"path": "/mnt/tank"})
		assertResultJSON(t, res)
	})

	t.Run("empty path returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "list_directory", map[string]any{"path": ""})
		assertError(t, res, "path must not be empty")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			listDirectoryFn: func(_ context.Context, _ string) ([]truenas.DirEntry, error) {
				return nil, errors.New("path not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_directory", map[string]any{"path": "/mnt/noexist"})
		assertError(t, res, "path not found")
	})
}
