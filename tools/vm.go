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
}
