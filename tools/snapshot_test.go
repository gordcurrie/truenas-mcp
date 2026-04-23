package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

func TestListSnapshots(t *testing.T) {
	t.Run("returns snapshots as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			listSnapshotsFn: func(_ context.Context, _ string, _ ...truenas.ListOptions) ([]truenas.Snapshot, error) {
				return []truenas.Snapshot{{ID: "tank/data@snap1"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_snapshots", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			listSnapshotsFn: func(_ context.Context, _ string, _ ...truenas.ListOptions) ([]truenas.Snapshot, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_snapshots", nil)
		assertError(t, res, "API error")
	})
}

func TestGetSnapshot(t *testing.T) {
	t.Run("returns snapshot as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			getSnapshotFn: func(_ context.Context, id string) (*truenas.Snapshot, error) {
				return &truenas.Snapshot{ID: id}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_snapshot", map[string]any{"id": "tank/data@snap1"})
		assertResultJSON(t, res)
	})

	t.Run("empty id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "get_snapshot", map[string]any{"id": ""})
		assertError(t, res, "id must not be empty")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			getSnapshotFn: func(_ context.Context, _ string) (*truenas.Snapshot, error) {
				return nil, errors.New("snapshot not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_snapshot", map[string]any{"id": "tank/data@noexist"})
		assertError(t, res, "snapshot not found")
	})
}

func TestCreateSnapshot(t *testing.T) {
	t.Run("returns created snapshot as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			createSnapshotFn: func(_ context.Context, p truenas.CreateSnapshotParams) (*truenas.Snapshot, error) {
				return &truenas.Snapshot{ID: p.Dataset + "@" + p.Name}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_snapshot", map[string]any{
			"dataset": "tank/data",
			"name":    "before-upgrade",
		})
		assertResultJSON(t, res)
	})

	t.Run("empty dataset returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "create_snapshot", map[string]any{"dataset": "", "name": "snap1"})
		assertError(t, res, "dataset must not be empty")
	})

	t.Run("empty name returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "create_snapshot", map[string]any{"dataset": "tank/data", "name": ""})
		assertError(t, res, "name must not be empty")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			createSnapshotFn: func(_ context.Context, _ truenas.CreateSnapshotParams) (*truenas.Snapshot, error) {
				return nil, errors.New("dataset busy")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_snapshot", map[string]any{"dataset": "tank/data", "name": "snap1"})
		assertError(t, res, "dataset busy")
	})
}
