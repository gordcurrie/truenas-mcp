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
		Name string `json:"name" jsonschema:"required,App name (as shown in the TrueNAS UI)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_container",
		Description: "Get detailed information about a specific Docker container by its app name.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p getContainerInput) (*mcp.CallToolResult, any, error) {
		container, err := client.GetContainer(ctx, p.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("get_container: %w", err)
		}
		return jsonResult(container)
	})

	type startContainerInput struct {
		Name string `json:"name" jsonschema:"required,App name (as shown in the TrueNAS UI)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "start_container",
		Description: "Start a Docker container by its app name. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p startContainerInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.StartContainer(ctx, p.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("start_container: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type stopContainerInput struct {
		Name string `json:"name" jsonschema:"required,App name (as shown in the TrueNAS UI)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "stop_container",
		Description: "Stop a running Docker container by its app name. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p stopContainerInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.StopContainer(ctx, p.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("stop_container: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type restartContainerInput struct {
		Name string `json:"name" jsonschema:"required,App name (as shown in the TrueNAS UI)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "restart_container",
		Description: "Restart a Docker container by its app name. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p restartContainerInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.RestartContainer(ctx, p.Name)
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

	type createContainerInput struct {
		AppName    string `json:"app_name"         jsonschema:"required,Instance name for the new app (lowercase, hyphens allowed, max 40 chars)"`
		CatalogApp string `json:"catalog_app"      jsonschema:"required,Catalog app to install (e.g. jellyfin)"`
		Train      string `json:"train,omitempty"  jsonschema:"Catalog train (default: stable)"`
		Version    string `json:"version,omitempty" jsonschema:"App version to install (default: latest)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_container",
		Description: "Install a catalog app from the TrueNAS app catalog. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p createContainerInput) (*mcp.CallToolResult, any, error) {
		train := p.Train
		if train == "" {
			train = "stable"
		}
		version := p.Version
		if version == "" {
			version = "latest"
		}
		jobID, err := client.CreateContainer(ctx, &truenas.CreateContainerParams{
			AppName:    p.AppName,
			CatalogApp: p.CatalogApp,
			Train:      train,
			Version:    version,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create_container: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type createCustomContainerInput struct {
		AppName                   string `json:"app_name"                    jsonschema:"required,Instance name for the new app (lowercase, hyphens allowed, max 40 chars)"`
		CustomComposeConfigString string `json:"custom_compose_config_string" jsonschema:"required,Docker Compose YAML defining the services to deploy"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_custom_container",
		Description: "Install a custom Docker Compose app on TrueNAS SCALE. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p createCustomContainerInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.CreateContainer(ctx, &truenas.CreateContainerParams{
			AppName:                   p.AppName,
			CustomApp:                 true,
			CustomComposeConfigString: p.CustomComposeConfigString,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create_custom_container: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})
}
