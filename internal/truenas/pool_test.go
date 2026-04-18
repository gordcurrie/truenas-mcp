package truenas

import (
	"context"
	"encoding/json"
	"errors"
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

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.query": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
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

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.get_instance": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
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

	srv := wsTestServer(t, map[string]methodHandler{
		"pool.get_instance": func(_ json.RawMessage) (any, *rpcError) {
			return nil, &rpcError{Code: -32001, Message: "not found"}
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetPool(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
