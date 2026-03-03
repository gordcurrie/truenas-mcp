package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListVMs_success(t *testing.T) {
	t.Parallel()

	want := []VM{
		{
			ID:     1,
			Name:   "pbs",
			VCPUs:  2,
			Memory: 4096,
			Status: VMStatus{State: "STOPPED"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vm" {
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
	got, err := c.ListVMs(context.Background())
	if err != nil {
		t.Fatalf("ListVMs: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 VM, got %d", len(got))
	}
	if got[0].Name != want[0].Name {
		t.Errorf("Name = %q, want %q", got[0].Name, want[0].Name)
	}
	if got[0].Memory != want[0].Memory {
		t.Errorf("Memory = %d, want %d", got[0].Memory, want[0].Memory)
	}
}

func TestGetVM_success(t *testing.T) {
	t.Parallel()

	want := VM{
		ID:     1,
		Name:   "pbs",
		VCPUs:  2,
		Memory: 4096,
		Status: VMStatus{State: "RUNNING"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vm/id/1" {
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
	got, err := c.GetVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.Status.State != want.Status.State {
		t.Errorf("Status.State = %q, want %q", got.Status.State, want.Status.State)
	}
}

func TestGetVM_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetVM(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStartVM_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/vm/id/1/start" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(42); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.StartVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("StartVM: %v", err)
	}
	if jobID != 42 {
		t.Errorf("jobID = %d, want 42", jobID)
	}
}

func TestStopVM_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/vm/id/1/stop" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(43); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.StopVM(context.Background(), 1, false)
	if err != nil {
		t.Fatalf("StopVM: %v", err)
	}
	if jobID != 43 {
		t.Errorf("jobID = %d, want 43", jobID)
	}
}

func TestRestartVM_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/vm/id/1/restart" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(44); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.RestartVM(context.Background(), 1)
	if err != nil {
		t.Fatalf("RestartVM: %v", err)
	}
	if jobID != 44 {
		t.Errorf("jobID = %d, want 44", jobID)
	}
}

func TestCreateVM_success(t *testing.T) {
	t.Parallel()

	created := VM{
		ID:     2,
		Name:   "pbs",
		VCPUs:  2,
		Memory: 4096,
		Status: VMStatus{State: "STOPPED"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/vm" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(created); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.CreateVM(context.Background(), &CreateVMParams{
		Name:   "pbs",
		VCPUs:  2,
		Memory: 4096,
	})
	if err != nil {
		t.Fatalf("CreateVM: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
	if got.Name != created.Name {
		t.Errorf("Name = %q, want %q", got.Name, created.Name)
	}
}

func TestCreateVM_validationErrors(t *testing.T) {
	t.Parallel()

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: test-api-key is a fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name   string
		params CreateVMParams
	}{
		{"empty name", CreateVMParams{Memory: 4096}},
		{"zero memory", CreateVMParams{Name: "pbs"}},
		{"invalid name with hyphen", CreateVMParams{Name: "test-vm", Memory: 4096}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.CreateVM(context.Background(), &tt.params)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestUpdateVM_success(t *testing.T) {
	t.Parallel()

	updated := VM{
		ID:     1,
		Name:   "pbs",
		VCPUs:  4,
		Memory: 8192,
		Status: VMStatus{State: "STOPPED"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/vm/id/1" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(updated); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.UpdateVM(context.Background(), 1, &UpdateVMParams{VCPUs: 4, Memory: 8192})
	if err != nil {
		t.Fatalf("UpdateVM: %v", err)
	}
	if got.VCPUs != updated.VCPUs {
		t.Errorf("VCPUs = %d, want %d", got.VCPUs, updated.VCPUs)
	}
	if got.Memory != updated.Memory {
		t.Errorf("Memory = %d, want %d", got.Memory, updated.Memory)
	}
}

func TestUpdateVM_validationErrors(t *testing.T) {
	t.Parallel()

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: test-api-key is a fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name   string
		params UpdateVMParams
	}{
		{"invalid name with hyphen", UpdateVMParams{Name: "test-vm"}},
		{"negative memory", UpdateVMParams{Memory: -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.UpdateVM(context.Background(), 1, &tt.params)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestDeleteVM_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/vm/id/1" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteVM(context.Background(), 1); err != nil {
		t.Fatalf("DeleteVM: %v", err)
	}
}

func TestDeleteVM_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.DeleteVM(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListVMDevices_success(t *testing.T) {
	t.Parallel()

	all := []VMDevice{
		{ID: 1, VMID: 2, DType: VMDeviceTypeDISK, Attributes: map[string]any{"path": "zvol/Storage/pbs/disk0", "type": "VIRTIO"}},
		{ID: 2, VMID: 2, DType: VMDeviceTypeNIC, Attributes: map[string]any{"type": "VIRTIO", "nic_attach": "br0"}},
		{ID: 3, VMID: 9, DType: VMDeviceTypeCDROM, Attributes: map[string]any{"path": "/mnt/Storage/isos/other.iso"}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vm/device" {
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
	got, err := c.ListVMDevices(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListVMDevices: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 devices for VM 2, got %d", len(got))
	}
	for _, d := range got {
		if d.VMID != 2 {
			t.Errorf("unexpected VMID %d in results", d.VMID)
		}
	}
}

func TestListVMDevices_empty(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vm/device" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]VMDevice{}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.ListVMDevices(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListVMDevices: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 devices, got %d", len(got))
	}
}

func TestAddVMDevice_success(t *testing.T) {
	t.Parallel()

	want := VMDevice{
		ID:         5,
		VMID:       2,
		DType:      VMDeviceTypeCDROM,
		Attributes: map[string]any{"path": "/mnt/Storage/isos/pbs.iso"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/vm/device" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Assert wire format: dtype must be inside attributes, not top-level.
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, topLevel := body["dtype"]; topLevel {
			t.Errorf("dtype must not appear as a top-level field in the request body")
		}
		var attrs map[string]any
		if err := json.Unmarshal(body["attributes"], &attrs); err != nil {
			t.Errorf("decoding attributes: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if attrs["dtype"] != string(VMDeviceTypeCDROM) {
			t.Errorf("attributes.dtype = %v, want %q", attrs["dtype"], VMDeviceTypeCDROM)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(want); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.AddVMDevice(context.Background(), &AddVMDeviceParams{
		VMID:       2,
		DType:      VMDeviceTypeCDROM,
		Attributes: map[string]any{"path": "/mnt/Storage/isos/pbs.iso"},
	})
	if err != nil {
		t.Fatalf("AddVMDevice: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %d, want %d", got.ID, want.ID)
	}
	if got.DType != want.DType {
		t.Errorf("DType = %q, want %q", got.DType, want.DType)
	}
}

func TestAddVMDevice_validation(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, "http://localhost:9")

	tests := []struct {
		name   string
		params *AddVMDeviceParams
	}{
		{"nil params", nil},
		{"zero vmid", &AddVMDeviceParams{VMID: 0, DType: VMDeviceTypeDISK, Attributes: map[string]any{"path": "zvol/x"}}},
		{"empty dtype", &AddVMDeviceParams{VMID: 1, DType: "", Attributes: map[string]any{"path": "zvol/x"}}},
		{"empty attributes", &AddVMDeviceParams{VMID: 1, DType: VMDeviceTypeDISK, Attributes: map[string]any{}}},
		{"dtype in attributes", &AddVMDeviceParams{VMID: 1, DType: VMDeviceTypeDISK, Attributes: map[string]any{"dtype": "DISK", "path": "zvol/x"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.AddVMDevice(context.Background(), tt.params)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestDeleteVMDevice_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/vm/device/id/5" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteVMDevice(context.Background(), 5); err != nil {
		t.Fatalf("DeleteVMDevice: %v", err)
	}
}

func TestDeleteVMDevice_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.DeleteVMDevice(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
