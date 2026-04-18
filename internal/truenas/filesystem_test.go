package truenas

import (
	"context"
	"encoding/json"
	"testing"
)

func TestListDirectory_success(t *testing.T) {
	t.Parallel()

	want := []DirEntry{
		{Name: "proxmox-backup-server.iso", Path: "/mnt/Storage/pbs/proxmox-backup-server.iso", Type: "FILE", Size: 1073741824},
		{Name: "backups", Path: "/mnt/Storage/pbs/backups", Type: "DIRECTORY", Size: 4096},
	}

	srv := wsTestServer(t, map[string]methodHandler{
		"filesystem.listdir": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.ListDirectory(context.Background(), "/mnt/Storage/pbs")
	if err != nil {
		t.Fatalf("ListDirectory: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d entries, got %d", len(want), len(got))
	}
	if got[0].Name != want[0].Name {
		t.Errorf("Name = %q, want %q", got[0].Name, want[0].Name)
	}
	if got[0].Type != want[0].Type {
		t.Errorf("Type = %q, want %q", got[0].Type, want[0].Type)
	}
}

func TestListDirectory_emptyPath(t *testing.T) {
	t.Parallel()

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.ListDirectory(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

func TestListDirectory_relativePath(t *testing.T) {
	t.Parallel()

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.ListDirectory(context.Background(), "mnt/Storage/pbs")
	if err == nil {
		t.Error("expected error for relative path, got nil")
	}
}

func TestListDirectory_serverError(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"filesystem.listdir": func(_ json.RawMessage) (any, *rpcError) {
			return nil, &rpcError{Code: -32000, Message: "internal error"}
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.ListDirectory(context.Background(), "/mnt/Storage/pbs")
	if err == nil {
		t.Error("expected error on server error, got nil")
	}
}
