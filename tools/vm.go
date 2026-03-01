package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

// registerVMTools registers all VM-related MCP tools onto the server.
func registerVMTools(s *mcp.Server, client *truenas.Client) {
	type listVMsInput struct{}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_vms",
		Description: "List all virtual machines configured on the TrueNAS SCALE system, including their state (RUNNING/STOPPED), CPU, and memory.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listVMsInput) (*mcp.CallToolResult, any, error) {
		vms, err := client.ListVMs(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_vms: %w", err)
		}
		return jsonResult(vms)
	})

	type getVMInput struct {
		ID int `json:"id" jsonschema:"Numeric VM ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_vm",
		Description: "Get detailed information about a specific virtual machine by its numeric ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p getVMInput) (*mcp.CallToolResult, any, error) {
		vm, err := client.GetVM(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_vm: %w", err)
		}
		return jsonResult(vm)
	})

	type startVMInput struct {
		ID int `json:"id" jsonschema:"Numeric VM ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "start_vm",
		Description: "Start a virtual machine by its numeric ID. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p startVMInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.StartVM(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("start_vm: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type stopVMInput struct {
		ID    int  `json:"id" jsonschema:"Numeric VM ID"`
		Force bool `json:"force,omitempty" jsonschema:"Force-terminate the VM without a graceful shutdown; defaults to false"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "stop_vm",
		Description: "Stop a virtual machine by its numeric ID. Set force=true to forcibly terminate without a graceful shutdown. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p stopVMInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.StopVM(ctx, p.ID, p.Force)
		if err != nil {
			return nil, nil, fmt.Errorf("stop_vm: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type restartVMInput struct {
		ID int `json:"id" jsonschema:"Numeric VM ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "restart_vm",
		Description: "Restart a virtual machine by its numeric ID. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p restartVMInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.RestartVM(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("restart_vm: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type createVMInput struct {
		Name            string `json:"name"                       jsonschema:"required,VM name (alphanumeric characters only; hyphens and special characters are not allowed)"`
		Memory          int    `json:"memory"                     jsonschema:"required,RAM in MiB (e.g. 4096 for 4 GiB)"`
		Description     string `json:"description,omitempty"      jsonschema:"Optional description"`
		VCPUs           int    `json:"vcpus,omitempty"            jsonschema:"Total virtual CPU count; defaults to 1"`
		Bootloader      string `json:"bootloader,omitempty"       jsonschema:"UEFI (default), UEFI_CSM, or GRUB"`
		Autostart       bool   `json:"autostart,omitempty"        jsonschema:"Start VM automatically with the system"`
		Cores           int    `json:"cores,omitempty"            jsonschema:"Cores per socket"`
		Threads         int    `json:"threads,omitempty"          jsonschema:"Threads per core"`
		ShutdownTimeout int    `json:"shutdown_timeout,omitempty" jsonschema:"Seconds to wait before force-killing on shutdown"`
		CPUMode         string `json:"cpu_mode,omitempty"         jsonschema:"CUSTOM, HOST-MODEL, or HOST-PASSTHROUGH"`
		CPUModel        string `json:"cpu_model,omitempty"        jsonschema:"CPU model name; only used when cpu_mode=CUSTOM"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_vm",
		Description: "Create a new virtual machine. At minimum provide name and memory (in MiB). Returns the created VM. Note: VM names must be alphanumeric only — hyphens, underscores, and other special characters are not allowed.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p createVMInput) (*mcp.CallToolResult, any, error) {
		vm, err := client.CreateVM(ctx, &truenas.CreateVMParams{
			Name:            p.Name,
			Memory:          p.Memory,
			Description:     p.Description,
			VCPUs:           p.VCPUs,
			Bootloader:      p.Bootloader,
			Autostart:       p.Autostart,
			Cores:           p.Cores,
			Threads:         p.Threads,
			ShutdownTimeout: p.ShutdownTimeout,
			CPUMode:         p.CPUMode,
			CPUModel:        p.CPUModel,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create_vm: %w", err)
		}
		return jsonResult(vm)
	})

	type updateVMInput struct {
		ID              int    `json:"id"                         jsonschema:"required,Numeric VM ID"`
		Name            string `json:"name,omitempty"             jsonschema:"New VM name"`
		Description     string `json:"description,omitempty"      jsonschema:"New description"`
		VCPUs           int    `json:"vcpus,omitempty"            jsonschema:"New total virtual CPU count"`
		Memory          int    `json:"memory,omitempty"           jsonschema:"New RAM in MiB"`
		Bootloader      string `json:"bootloader,omitempty"       jsonschema:"UEFI, UEFI_CSM, or GRUB"`
		Cores           int    `json:"cores,omitempty"            jsonschema:"Cores per socket"`
		Threads         int    `json:"threads,omitempty"          jsonschema:"Threads per core"`
		ShutdownTimeout int    `json:"shutdown_timeout,omitempty" jsonschema:"Seconds before force-kill on shutdown"`
		CPUMode         string `json:"cpu_mode,omitempty"         jsonschema:"CUSTOM, HOST-MODEL, or HOST-PASSTHROUGH"`
		CPUModel        string `json:"cpu_model,omitempty"        jsonschema:"CPU model; only used when cpu_mode=CUSTOM"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_vm",
		Description: "Update configuration of an existing VM by ID. Only fields with non-zero/non-empty values are applied; fields left unset or set to their zero value are ignored and remain unchanged. Clearing a field to empty string or 0 is not supported. Returns the updated VM.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p updateVMInput) (*mcp.CallToolResult, any, error) {
		vm, err := client.UpdateVM(ctx, p.ID, &truenas.UpdateVMParams{
			Name:            p.Name,
			Description:     p.Description,
			VCPUs:           p.VCPUs,
			Memory:          p.Memory,
			Bootloader:      p.Bootloader,
			Cores:           p.Cores,
			Threads:         p.Threads,
			ShutdownTimeout: p.ShutdownTimeout,
			CPUMode:         p.CPUMode,
			CPUModel:        p.CPUModel,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("update_vm: %w", err)
		}
		return jsonResult(vm)
	})

	type listVMDevicesInput struct {
		ID int `json:"id" jsonschema:"required,Numeric VM ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_vm_devices",
		Description: "List all hardware devices attached to a VM (disks, CDROMs, NICs, displays). Returns each device's ID, type, and attributes.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p listVMDevicesInput) (*mcp.CallToolResult, any, error) {
		devices, err := client.ListVMDevices(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("list_vm_devices: %w", err)
		}
		return jsonResult(devices)
	})

	type addVMDeviceInput struct {
		VMID       int            `json:"vm_id"    jsonschema:"required,Numeric VM ID to attach the device to"`
		DType      string         `json:"dtype"    jsonschema:"required,Device type: DISK | CDROM | NIC | DISPLAY. DISK attrs: {path:'zvol/pool/dataset' type:'VIRTIO'}. CDROM attrs: {path:'/mnt/Storage/isos/file.iso'}. NIC attrs: {type:'VIRTIO' nic_attach:'br0'}. DISPLAY attrs: {web:true port:5900}"`
		Attributes map[string]any `json:"attributes" jsonschema:"required,Device-type-specific key-value pairs. See dtype description for examples."`
		Order      *int           `json:"order,omitempty" jsonschema:"Optional boot/device order index"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_vm_device",
		Description: "Attach a hardware device to a VM. Supports DISK (zvol-backed), CDROM (ISO file), NIC (virtio/e1000), and DISPLAY (VNC). The VM does not need to be stopped to add devices, but changes take effect on next boot.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p addVMDeviceInput) (*mcp.CallToolResult, any, error) {
		device, err := client.AddVMDevice(ctx, &truenas.AddVMDeviceParams{
			VMID:       p.VMID,
			DType:      truenas.VMDeviceType(p.DType),
			Attributes: p.Attributes,
			Order:      p.Order,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("add_vm_device: %w", err)
		}
		return jsonResult(device)
	})

	type deleteVMDeviceInput struct {
		ID int `json:"id" jsonschema:"required,Numeric device ID (from list_vm_devices)"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_vm_device",
		Description: "Remove a hardware device from a VM by its device ID. Use list_vm_devices to find device IDs. Changes take effect on next boot.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p deleteVMDeviceInput) (*mcp.CallToolResult, any, error) {
		if err := client.DeleteVMDevice(ctx, p.ID); err != nil {
			return nil, nil, fmt.Errorf("delete_vm_device: %w", err)
		}
		return jsonResult(map[string]any{"deleted": true, "device_id": p.ID})
	})
}
