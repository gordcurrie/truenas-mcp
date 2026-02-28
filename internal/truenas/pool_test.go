package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPools_success(t *testing.T) {
	t.Parallel()

	want := []Pool{
		{
			ID:      1,
			Name:    "Storage",
			GUID:    "1234567890",
			Status:  "ONLINE",
			Path:    "/mnt/Storage",
			Healthy: true,
			Size:    31988916420608,
			Free:    27102360289280,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pool" {
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
	got, err := c.ListPools(context.Background())
	if err != nil {
		t.Fatalf("ListPools: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 pool, got %d", len(got))
	}
	if got[0].Name != want[0].Name {
		t.Errorf("Name = %q, want %q", got[0].Name, want[0].Name)
	}
	if got[0].Free != want[0].Free {
		t.Errorf("Free = %d, want %d", got[0].Free, want[0].Free)
	}
}

func TestGetPool_success(t *testing.T) {
	t.Parallel()

	want := Pool{
		ID:      1,
		Name:    "Storage",
		Status:  "ONLINE",
		Healthy: true,
		Size:    31988916420608,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pool/id/1" {
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
	got, err := c.GetPool(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetPool: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.Size != want.Size {
		t.Errorf("Size = %d, want %d", got.Size, want.Size)
	}
}

func TestGetPool_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetPool(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
