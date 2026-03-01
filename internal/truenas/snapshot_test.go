package truenas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pool/snapshot" {
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

	all := []Snapshot{
		{ID: "Storage/backups@snap1", Dataset: "Storage/backups", Name: "snap1", Pool: "Storage"},
		{ID: "apps/data@snap1", Dataset: "apps/data", Name: "snap1", Pool: "apps"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pool/snapshot" {
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Go's HTTP server decodes %2F to / in r.URL.Path; check the decoded form.
		if r.URL.Path != "/pool/snapshot/id/Storage/backups@before-upgrade" {
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

	c := newTestClient(t, "http://localhost")

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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/pool/snapshot" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var body CreateSnapshotParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		if body.Dataset != "Storage/backups" {
			t.Errorf("Dataset = %q, want %q", body.Dataset, "Storage/backups")
		}
		if body.Name != "before-upgrade" {
			t.Errorf("Name = %q, want %q", body.Name, "before-upgrade")
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(want); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
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

	c := newTestClient(t, "http://localhost")

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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/pool/snapshot/id/Storage/backups@before-upgrade/rollback" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(nil); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.RollbackSnapshot(context.Background(), "Storage/backups@before-upgrade", RollbackSnapshotParams{}); err != nil {
		t.Fatalf("RollbackSnapshot: %v", err)
	}
}

func TestRollbackSnapshot_validation(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, "http://localhost")

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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/pool/snapshot/id/Storage/backups@before-upgrade" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(true); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteSnapshot(context.Background(), "Storage/backups@before-upgrade"); err != nil {
		t.Fatalf("DeleteSnapshot: %v", err)
	}
}

func TestDeleteSnapshot_validation(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, "http://localhost")

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
