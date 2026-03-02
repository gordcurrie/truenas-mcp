package truenas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListInterfaces_success(t *testing.T) {
	t.Parallel()

	want := []Interface{
		{ID: "br0", Name: "br0", Type: "BRIDGE", Description: ""},
		{ID: "enp1s0", Name: "enp1s0", Type: "PHYSICAL", Description: ""},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/interface" {
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
	got, err := c.ListInterfaces(context.Background())
	if err != nil {
		t.Fatalf("ListInterfaces: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d interfaces, got %d", len(want), len(got))
	}
	if got[0].ID != want[0].ID {
		t.Errorf("ID = %q, want %q", got[0].ID, want[0].ID)
	}
	if got[0].Type != want[0].Type {
		t.Errorf("Type = %q, want %q", got[0].Type, want[0].Type)
	}
}

func TestListInterfaces_serverError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.ListInterfaces(context.Background())
	if err == nil {
		t.Error("expected error on server 500, got nil")
	}
}
