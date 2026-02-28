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
		Name            string `json:"name"                       jsonschema:"required,VM name"`
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
		Description: "Create a new virtual machine. At minimum provide name and memory (in MiB). Returns the created VM.",
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
		Description: "Update configuration of an existing VM by ID. Only supplied fields are changed; omitted fields are left as-is. Returns the updated VM.",
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
}
