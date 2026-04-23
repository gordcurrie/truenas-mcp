package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

func TestListVMs(t *testing.T) {
	t.Run("returns VMs as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			listVMsFn: func(_ context.Context, _ ...truenas.ListOptions) ([]truenas.VM, error) {
				return []truenas.VM{{ID: 1, Name: "pbs"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_vms", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			listVMsFn: func(_ context.Context, _ ...truenas.ListOptions) ([]truenas.VM, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_vms", nil)
		assertError(t, res, "API error")
	})
}

func TestGetVM(t *testing.T) {
	t.Run("returns VM as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			getVMFn: func(_ context.Context, id int) (*truenas.VM, error) {
				return &truenas.VM{ID: id, Name: "pbs"}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_vm", map[string]any{"id": 1})
		assertResultJSON(t, res)
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "get_vm", map[string]any{"id": 0})
		assertError(t, res, "id must be a positive integer")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			getVMFn: func(_ context.Context, _ int) (*truenas.VM, error) {
				return nil, errors.New("VM not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_vm", map[string]any{"id": 99})
		assertError(t, res, "VM not found")
	})
}

func TestStartVM(t *testing.T) {
	t.Run("returns job ID as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			startVMFn: func(_ context.Context, _ int) (int, error) {
				return 42, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "start_vm", map[string]any{"id": 1})
		assertResultJSON(t, res)
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "start_vm", map[string]any{"id": 0})
		assertError(t, res, "id must be a positive integer")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockTruenasClient{
			startVMFn: func(_ context.Context, _ int) (int, error) {
				return 0, errors.New("VM already running")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "start_vm", map[string]any{"id": 1})
		assertError(t, res, "VM already running")
	})
}

func TestStopVM(t *testing.T) {
	t.Run("returns job ID as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			stopVMFn: func(_ context.Context, _ int, _ bool) (int, error) {
				return 43, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "stop_vm", map[string]any{"id": 1})
		assertResultJSON(t, res)
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "stop_vm", map[string]any{"id": -1})
		assertError(t, res, "id must be a positive integer")
	})
}

func TestRestartVM(t *testing.T) {
	t.Run("returns job ID as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			restartVMFn: func(_ context.Context, _ int) (int, error) {
				return 44, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "restart_vm", map[string]any{"id": 1})
		assertResultJSON(t, res)
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "restart_vm", map[string]any{"id": 0})
		assertError(t, res, "id must be a positive integer")
	})
}

func TestCreateVM(t *testing.T) {
	t.Run("returns created VM as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			createVMFn: func(_ context.Context, p *truenas.CreateVMParams) (*truenas.VM, error) {
				return &truenas.VM{ID: 1, Name: p.Name}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_vm", map[string]any{"name": "pbs", "memory": 4096})
		assertResultJSON(t, res)
	})

	t.Run("empty name returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "create_vm", map[string]any{"name": "", "memory": 4096})
		assertError(t, res, "name must not be empty")
	})

	t.Run("zero memory returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "create_vm", map[string]any{"name": "pbs", "memory": 0})
		assertError(t, res, "memory must be greater than 0")
	})
}

func TestUpdateVM(t *testing.T) {
	t.Run("returns updated VM as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			updateVMFn: func(_ context.Context, id int, _ *truenas.UpdateVMParams) (*truenas.VM, error) {
				return &truenas.VM{ID: id}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "update_vm", map[string]any{"id": 1, "memory": 8192})
		assertResultJSON(t, res)
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "update_vm", map[string]any{"id": 0})
		assertError(t, res, "id must be a positive integer")
	})
}

func TestListVMDevices(t *testing.T) {
	t.Run("returns devices as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			listVMDevicesFn: func(_ context.Context, _ int) ([]truenas.VMDevice, error) {
				return []truenas.VMDevice{{ID: 1, VMID: 1}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_vm_devices", map[string]any{"id": 1})
		assertResultJSON(t, res)
	})

	t.Run("invalid id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "list_vm_devices", map[string]any{"id": 0})
		assertError(t, res, "id must be a positive integer")
	})
}

func TestAddVMDevice(t *testing.T) {
	t.Run("returns added device as JSON", func(t *testing.T) {
		mock := &mockTruenasClient{
			addVMDeviceFn: func(_ context.Context, p *truenas.AddVMDeviceParams) (*truenas.VMDevice, error) {
				return &truenas.VMDevice{ID: 10, VMID: p.VMID}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "add_vm_device", map[string]any{
			"vm_id":      1,
			"dtype":      "NIC",
			"attributes": map[string]any{"type": "VIRTIO", "nic_attach": "br0"},
		})
		assertResultJSON(t, res)
	})

	t.Run("invalid vm_id returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "add_vm_device", map[string]any{
			"vm_id":      0,
			"dtype":      "NIC",
			"attributes": map[string]any{},
		})
		assertError(t, res, "vm_id must be a positive integer")
	})

	t.Run("invalid dtype returns error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockTruenasClient{})
		defer cleanup()

		res := callTool(t, cs, "add_vm_device", map[string]any{
			"vm_id":      1,
			"dtype":      "INVALID",
			"attributes": map[string]any{},
		})
		assertError(t, res, "dtype must be one of")
	})
}
