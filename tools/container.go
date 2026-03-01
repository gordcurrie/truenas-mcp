package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerContainerTools adds Docker container and image MCP tools to the server.
func registerContainerTools(s *mcp.Server, client *truenas.Client) {
	type listContainersInput struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_containers",
		Description: "List all Docker containers managed by TrueNAS SCALE.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listContainersInput) (*mcp.CallToolResult, any, error) {
		containers, err := client.ListContainers(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_containers: %w", err)
		}
		return jsonResult(containers)
	})

	type getContainerInput struct {
		ID int `json:"id" jsonschema:"required,Numeric container ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_container",
		Description: "Get detailed information about a specific Docker container by its numeric ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p getContainerInput) (*mcp.CallToolResult, any, error) {
		container, err := client.GetContainer(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_container: %w", err)
		}
		return jsonResult(container)
	})

	type startContainerInput struct {
		ID int `json:"id" jsonschema:"required,Numeric container ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "start_container",
		Description: "Start a Docker container by its numeric ID. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p startContainerInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.StartContainer(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("start_container: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type stopContainerInput struct {
		ID int `json:"id" jsonschema:"required,Numeric container ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "stop_container",
		Description: "Stop a running Docker container by its numeric ID. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p stopContainerInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.StopContainer(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("stop_container: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type restartContainerInput struct {
		ID int `json:"id" jsonschema:"required,Numeric container ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "restart_container",
		Description: "Restart a Docker container by its numeric ID. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p restartContainerInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.RestartContainer(ctx, p.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("restart_container: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type listImagesInput struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_images",
		Description: "List all Docker images stored on the TrueNAS SCALE system.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listImagesInput) (*mcp.CallToolResult, any, error) {
		images, err := client.ListImages(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_images: %w", err)
		}
		return jsonResult(images)
	})
}
