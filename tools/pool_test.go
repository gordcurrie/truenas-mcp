package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

func TestListPools(t *testing.T) {
	t.Run("returns pools as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			listPoolsFn: func(_ context.Context, _ ...truenas.ListOptions) ([]truenas.Pool, error) {
				return []truenas.Pool{{ID: 1, Name: "tank"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_pools", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			listPoolsFn: func(_ context.Context, _ ...truenas.ListOptions) ([]truenas.Pool, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_pools", nil)
		assertError(t, res, "API error")
	})
}

func TestGetPool(t *testing.T) {
	t.Run("returns pool as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			getPoolFn: func(_ context.Context, id int) (*truenas.Pool, error) {
				return &truenas.Pool{ID: id, Name: "tank"}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_pool", map[string]any{"id": 1})
		assertResultJSON(t, res)
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "get_pool", map[string]any{"id": 0})
		assertError(t, res, "id must be a positive integer")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			getPoolFn: func(_ context.Context, _ int) (*truenas.Pool, error) {
				return nil, errors.New("pool not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_pool", map[string]any{"id": 99})
		assertError(t, res, "pool not found")
	})
}
