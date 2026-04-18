package truenas

import (
	"context"
	"encoding/json"
	"errors"
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

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.dataset.query": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
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

	want := []Dataset{
		{ID: "Storage/backups", Name: "backups", Pool: "Storage", Type: "FILESYSTEM"},
	}

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.dataset.query": func(params json.RawMessage) (any, *rpcError) {
			// params is [[filters], {options}] — check the filter contains pool=Storage
			var p []json.RawMessage
			if err := json.Unmarshal(params, &p); err != nil || len(p) < 1 {
				return nil, &rpcError{Code: -32600, Message: "bad params"}
			}
			var filters [][]any
			if err := json.Unmarshal(p[0], &filters); err != nil || len(filters) == 0 {
				return nil, &rpcError{Code: -32600, Message: "expected filters"}
			}
			// filters[0] should be ["pool", "=", "Storage"]
			if len(filters[0]) != 3 || filters[0][0] != "pool" || filters[0][2] != "Storage" {
				return nil, &rpcError{Code: -32600, Message: "wrong filter"}
			}
			return want, nil
		},
	})
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

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.dataset.get_instance": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
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

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.dataset.get_instance": func(_ json.RawMessage) (any, *rpcError) {
			return nil, &rpcError{Code: -32001, Message: "not found"}
		},
	})
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

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.dataset.create": func(_ json.RawMessage) (any, *rpcError) {
			return created, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.CreateDataset(context.Background(), &CreateDatasetParams{
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

	// No server needed — validation fires before any network call.
	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.CreateDataset(context.Background(), &CreateDatasetParams{})
	if err == nil {
		t.Error("expected error for empty Name, got nil")
	}
}

func TestCreateDataset_nilParams(t *testing.T) {
	t.Parallel()

	// No server needed — nil guard fires before any network call.
	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.CreateDataset(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil params, got nil")
	}
}

func TestCreateDataset_zvolMissingVolsize(t *testing.T) {
	t.Parallel()

	// No server needed — validation fires before any network call.
	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.CreateDataset(context.Background(), &CreateDatasetParams{
		Name: "Storage/vm-disk",
		Type: "VOLUME",
	})
	if err == nil {
		t.Error("expected error for VOLUME type without volsize, got nil")
	}
}
