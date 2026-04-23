package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

func TestListDatasets(t *testing.T) {
	t.Run("returns datasets as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			listDatasetsFn: func(_ context.Context, _ string, _ ...truenas.ListOptions) ([]truenas.Dataset, error) {
				return []truenas.Dataset{{ID: "tank/data"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_datasets", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			listDatasetsFn: func(_ context.Context, _ string, _ ...truenas.ListOptions) ([]truenas.Dataset, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_datasets", nil)
		assertError(t, res, "API error")
	})
}

func TestGetDataset(t *testing.T) {
	t.Run("returns dataset as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			getDatasetFn: func(_ context.Context, id string) (*truenas.Dataset, error) {
				return &truenas.Dataset{ID: id}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_dataset", map[string]any{"id": "tank/data"})
		assertResultJSON(t, res)
	})

	t.Run("empty id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "get_dataset", map[string]any{"id": ""})
		assertError(t, res, "id must not be empty")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			getDatasetFn: func(_ context.Context, _ string) (*truenas.Dataset, error) {
				return nil, errors.New("dataset not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_dataset", map[string]any{"id": "tank/noexist"})
		assertError(t, res, "dataset not found")
	})
}

func TestCreateDataset(t *testing.T) {
	t.Run("returns created dataset as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			createDatasetFn: func(_ context.Context, p *truenas.CreateDatasetParams) (*truenas.Dataset, error) {
				return &truenas.Dataset{ID: p.Name}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_dataset", map[string]any{"name": "tank/backups"})
		assertResultJSON(t, res)
	})

	t.Run("empty name returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "create_dataset", map[string]any{"name": ""})
		assertError(t, res, "name must not be empty")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			createDatasetFn: func(_ context.Context, _ *truenas.CreateDatasetParams) (*truenas.Dataset, error) {
				return nil, errors.New("dataset already exists")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_dataset", map[string]any{"name": "tank/dup"})
		assertError(t, res, "dataset already exists")
	})
}
