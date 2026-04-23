package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

func TestListInterfaces(t *testing.T) {
	t.Run("returns interfaces as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			listInterfacesFn: func(_ context.Context) ([]truenas.Interface, error) {
				return []truenas.Interface{{Name: "eth0"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_interfaces", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			listInterfacesFn: func(_ context.Context) ([]truenas.Interface, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_interfaces", nil)
		assertError(t, res, "API error")
	})
}
