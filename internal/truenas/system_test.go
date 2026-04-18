package truenas

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	t.Parallel()

	want := SystemInfo{
		Version:       "TrueNAS-SCALE-26.0",
		Hostname:      "truenas",
		PhysMem:       17179869184,
		Model:         "QEMU Virtual CPU version 2.5+",
		Cores:         4,
		LoadAvg:       []float64{0.1, 0.2, 0.3},
		UptimeSeconds: 12345.6,
		Timezone:      "America/Vancouver",
		SystemProduct: "TrueNAS",
		Manufacturer:  "iXsystems",
	}

	srv := wsTestServer(t, map[string]methodHandler{
		"system.info": func(_ json.RawMessage) (any, *rpcError) {
			return want, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatalf("GetSystemInfo: %v", err)
	}

	if got.Version != want.Version {
		t.Errorf("Version = %q, want %q", got.Version, want.Version)
	}
	if got.Hostname != want.Hostname {
		t.Errorf("Hostname = %q, want %q", got.Hostname, want.Hostname)
	}
	if got.Cores != want.Cores {
		t.Errorf("Cores = %d, want %d", got.Cores, want.Cores)
	}
	if got.Timezone != want.Timezone {
		t.Errorf("Timezone = %q, want %q", got.Timezone, want.Timezone)
	}
}

func TestGetSystemInfo_serverError(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"system.info": func(_ json.RawMessage) (any, *rpcError) {
			return nil, &rpcError{Code: -32000, Message: "internal error"}
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetSystemInfo(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
