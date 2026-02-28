package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListDatasets_success(t *testing.T) {
	t.Parallel()

	want := []Dataset{
		{
			ID:   "Storage/backups",
			Name: "backups",
			Pool: "Storage",
			Type: "FILESYSTEM",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pool/dataset" {
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
	got, err := c.ListDatasets(context.Background(), "")
	if err != nil {
		t.Fatalf("ListDatasets: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 dataset, got %d", len(got))
	}
	if got[0].ID != want[0].ID {
		t.Errorf("ID = %q, want %q", got[0].ID, want[0].ID)
	}
}

func TestListDatasets_filterByPool(t *testing.T) {
	t.Parallel()

	all := []Dataset{
		{ID: "Storage/backups", Name: "backups", Pool: "Storage", Type: "FILESYSTEM"},
		{ID: "apps/data", Name: "data", Pool: "apps", Type: "FILESYSTEM"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pool/dataset" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(all); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.ListDatasets(context.Background(), "Storage")
	if err != nil {
		t.Fatalf("ListDatasets: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 filtered dataset, got %d", len(got))
	}
	if got[0].Pool != "Storage" {
		t.Errorf("Pool = %q, want %q", got[0].Pool, "Storage")
	}
}

func TestGetDataset_success(t *testing.T) {
	t.Parallel()

	want := Dataset{
		ID:   "Storage/backups",
		Name: "backups",
		Pool: "Storage",
		Type: "FILESYSTEM",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The Go HTTP server decodes the path into r.URL.Path, but r.URL.RawPath
		// preserves the original percent-encoding. Use EscapedPath() which returns
		// RawPath when it differs from Path, so we can verify the %2F is preserved.
		if r.URL.EscapedPath() != "/pool/dataset/id/Storage%2Fbackups" {
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
	got, err := c.GetDataset(context.Background(), "Storage/backups")
	if err != nil {
		t.Fatalf("GetDataset: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
	if got.Pool != want.Pool {
		t.Errorf("Pool = %q, want %q", got.Pool, want.Pool)
	}
}

func TestGetDataset_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetDataset(context.Background(), "Storage/doesnotexist")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateDataset_success(t *testing.T) {
	t.Parallel()

	created := Dataset{
		ID:   "Storage/backups",
		Name: "backups",
		Pool: "Storage",
		Type: "FILESYSTEM",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/pool/dataset" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(created); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.CreateDataset(context.Background(), CreateDatasetParams{
		Name: "Storage/backups",
		Type: "FILESYSTEM",
	})
	if err != nil {
		t.Fatalf("CreateDataset: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestCreateDataset_emptyName(t *testing.T) {
	t.Parallel()

	// No server needed — validation should fail before any HTTP call.
	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: test-api-key is a fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.CreateDataset(context.Background(), CreateDatasetParams{})
	if err == nil {
		t.Error("expected error for empty Name, got nil")
	}
}
