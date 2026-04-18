package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestListSnapshots_success(t *testing.T) {
	t.Parallel()

	want := []Snapshot{
		{
			ID:      "Storage/backups@before-upgrade",
			Dataset: "Storage/backups",
			Name:    "before-upgrade",
			Pool:    "Storage",
		},
	}

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.snapshot.query": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.ListSnapshots(context.Background(), "")
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(got))
	}
	if got[0].ID != want[0].ID {
		t.Errorf("ID = %q, want %q", got[0].ID, want[0].ID)
	}
}

func TestListSnapshots_filterByDataset(t *testing.T) {
	t.Parallel()

	want := []Snapshot{
		{ID: "Storage/backups@snap1", Dataset: "Storage/backups", Name: "snap1", Pool: "Storage"},
	}

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.snapshot.query": func(params json.RawMessage) (any, *rpcError) {
			var p []json.RawMessage
			if err := json.Unmarshal(params, &p); err != nil || len(p) < 1 {
				return nil, &rpcError{Code: -32600, Message: "bad params"}
			}
			var filters [][]any
			if err := json.Unmarshal(p[0], &filters); err != nil || len(filters) == 0 {
				return nil, &rpcError{Code: -32600, Message: "expected filters"}
			}
			if len(filters[0]) != 3 || filters[0][0] != "dataset" || filters[0][2] != "Storage/backups" {
				return nil, &rpcError{Code: -32600, Message: "wrong filter"}
			}
			return want, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.ListSnapshots(context.Background(), "Storage/backups")
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 filtered snapshot, got %d", len(got))
	}
	if got[0].Dataset != "Storage/backups" {
		t.Errorf("Dataset = %q, want %q", got[0].Dataset, "Storage/backups")
	}
}

func TestGetSnapshot_success(t *testing.T) {
	t.Parallel()

	want := Snapshot{
		ID:      "Storage/backups@before-upgrade",
		Dataset: "Storage/backups",
		Name:    "before-upgrade",
		Pool:    "Storage",
	}

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.snapshot.get_instance": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.GetSnapshot(context.Background(), "Storage/backups@before-upgrade")
	if err != nil {
		t.Fatalf("GetSnapshot: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
}

func TestGetSnapshot_validation(t *testing.T) {
	t.Parallel()

	// No server needed — validation fires before any network call.
	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name string
		id   string
	}{
		{"empty id", ""},
		{"missing @", "Storage/backups"},
		{"empty name", "Storage/backups@"},
		{"empty dataset", "@before-upgrade"},
		{"multiple @", "Storage@bad@name"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := c.GetSnapshot(context.Background(), tc.id); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestCreateSnapshot_success(t *testing.T) {
	t.Parallel()

	want := Snapshot{
		ID:      "Storage/backups@before-upgrade",
		Dataset: "Storage/backups",
		Name:    "before-upgrade",
		Pool:    "Storage",
	}

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.snapshot.create": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.CreateSnapshot(context.Background(), CreateSnapshotParams{
		Dataset: "Storage/backups",
		Name:    "before-upgrade",
	})
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
}

func TestCreateSnapshot_validation(t *testing.T) {
	t.Parallel()

	// No server needed — validation fires before any network call.
	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name   string
		params CreateSnapshotParams
	}{
		{"empty dataset", CreateSnapshotParams{Dataset: "", Name: "snap1"}},
		{"empty name", CreateSnapshotParams{Dataset: "Storage/backups", Name: ""}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := c.CreateSnapshot(context.Background(), tc.params); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestRollbackSnapshot_success(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.snapshot.rollback": func(_ json.RawMessage) (any, *rpcError) {
			return nil, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.RollbackSnapshot(context.Background(), "Storage/backups@before-upgrade", RollbackSnapshotParams{}); err != nil {
		t.Fatalf("RollbackSnapshot: %v", err)
	}
}

func TestRollbackSnapshot_validation(t *testing.T) {
	t.Parallel()

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name string
		id   string
	}{
		{"empty id", ""},
		{"missing @", "Storage/backups"},
		{"empty name", "Storage/backups@"},
		{"empty dataset", "@before-upgrade"},
		{"multiple @", "Storage@bad@name"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := c.RollbackSnapshot(context.Background(), tc.id, RollbackSnapshotParams{}); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDeleteSnapshot_success(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.snapshot.delete": func(_ json.RawMessage) (any, *rpcError) {
			return nil, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteSnapshot(context.Background(), "Storage/backups@before-upgrade"); err != nil {
		t.Fatalf("DeleteSnapshot: %v", err)
	}
}

func TestDeleteSnapshot_validation(t *testing.T) {
	t.Parallel()

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name string
		id   string
	}{
		{"empty id", ""},
		{"missing @", "Storage/backups"},
		{"empty name", "Storage/backups@"},
		{"empty dataset", "@before-upgrade"},
		{"multiple @", "Storage@bad@name"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := c.DeleteSnapshot(context.Background(), tc.id); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestGetSnapshot_notFound(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.snapshot.get_instance": func(_ json.RawMessage) (any, *rpcError) {
			return nil, &rpcError{Code: -32001, Message: "not found"}
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetSnapshot(context.Background(), "Storage/backups@missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
