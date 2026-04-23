package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

func TestDeleteVM(t *testing.T) {
	t.Run("succeeds with confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_vm", map[string]any{"id": 1, "confirmed": true})
		assertResultJSON(t, res)
	})

	t.Run("requires confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_vm", map[string]any{"id": 1, "confirmed": false})
		assertError(t, res, "confirmed must be true")
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_vm", map[string]any{"id": 0, "confirmed": true})
		assertError(t, res, "id must be a positive integer")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			deleteVMFn: func(_ context.Context, _ int) error {
				return errors.New("VM is running")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "delete_vm", map[string]any{"id": 1, "confirmed": true})
		assertError(t, res, "VM is running")
	})
}

func TestDeleteApp(t *testing.T) {
	t.Run("succeeds with confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_app", map[string]any{"name": "jellyfin", "confirmed": true})
		assertResultJSON(t, res)
	})

	t.Run("requires confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_app", map[string]any{"name": "jellyfin", "confirmed": false})
		assertError(t, res, "confirmed must be true")
	})

	t.Run("empty name returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_app", map[string]any{"name": "", "confirmed": true})
		assertError(t, res, "name must not be empty")
	})
}

func TestDeleteSnapshot(t *testing.T) {
	t.Run("succeeds with confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_snapshot", map[string]any{
			"id": "tank/data@snap1", "confirmed": true,
		})
		assertResultJSON(t, res)
	})

	t.Run("requires confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_snapshot", map[string]any{
			"id": "tank/data@snap1", "confirmed": false,
		})
		assertError(t, res, "confirmed must be true")
	})

	t.Run("empty id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_snapshot", map[string]any{"id": "", "confirmed": true})
		assertError(t, res, "id must not be empty")
	})
}

func TestDeleteVMDevice(t *testing.T) {
	t.Run("succeeds with confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_vm_device", map[string]any{"id": 5, "confirmed": true})
		assertResultJSON(t, res)
	})

	t.Run("requires confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_vm_device", map[string]any{"id": 5, "confirmed": false})
		assertError(t, res, "confirmed must be true")
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "delete_vm_device", map[string]any{"id": 0, "confirmed": true})
		assertError(t, res, "id must be a positive integer")
	})
}

func TestRollbackSnapshot(t *testing.T) {
	t.Run("succeeds with confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "rollback_snapshot", map[string]any{
			"id": "tank/data@snap1", "confirmed": true,
		})
		assertResultJSON(t, res)
	})

	t.Run("requires confirmed=true", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "rollback_snapshot", map[string]any{
			"id": "tank/data@snap1", "confirmed": false,
		})
		assertError(t, res, "confirmed must be true")
	})

	t.Run("empty id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "rollback_snapshot", map[string]any{"id": "", "confirmed": true})
		assertError(t, res, "id must not be empty")
	})

	t.Run("recursive_clones without recursive returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "rollback_snapshot", map[string]any{
			"id":               "tank/data@snap1",
			"confirmed":        true,
			"recursive_clones": true,
			"recursive":        false,
		})
		assertError(t, res, "recursive_clones requires recursive=true")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			rollbackSnapshotFn: func(_ context.Context, _ string, _ truenas.RollbackSnapshotParams) error {
				return errors.New("dataset busy")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "rollback_snapshot", map[string]any{
			"id": "tank/data@snap1", "confirmed": true,
		})
		assertError(t, res, "dataset busy")
	})
}
