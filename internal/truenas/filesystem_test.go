package truenas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListDirectory_success(t *testing.T) {
	t.Parallel()

	want := []DirEntry{
		{Name: "proxmox-backup-server.iso", Path: "/mnt/Storage/pbs/proxmox-backup-server.iso", Type: "FILE", Size: 1073741824},
		{Name: "backups", Path: "/mnt/Storage/pbs/backups", Type: "DIRECTORY", Size: 4096},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/filesystem/listdir" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(want); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
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

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: test-api-key is a fake placeholder, not a real credential
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

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: test-api-key is a fake placeholder, not a real credential
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.ListDirectory(context.Background(), "/mnt/Storage/pbs")
	if err == nil {
		t.Error("expected error on server 500, got nil")
	}
}
