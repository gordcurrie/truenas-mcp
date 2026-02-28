package truenas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	t.Parallel()

	want := SystemInfo{
		Version:       "TrueNAS-SCALE-24.10.2",
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/system/info" {
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal error"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetSystemInfo(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
