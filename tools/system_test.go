package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

func TestGetSystemInfo(t *testing.T) {
	t.Run("returns system info as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			getSystemInfoFn: func(context.Context) (*truenas.SystemInfo, error) {
				return &truenas.SystemInfo{Hostname: "truenas", Version: "TrueNAS-SCALE-24.04"}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_system_info", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			getSystemInfoFn: func(context.Context) (*truenas.SystemInfo, error) {
				return nil, errors.New("connection refused")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_system_info", nil)
		assertError(t, res, "connection refused")
	})
}
