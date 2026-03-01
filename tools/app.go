package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerAppTools adds TrueNAS app and Docker image MCP tools to the server.
func registerAppTools(s *mcp.Server, client *truenas.Client) {
	type listAppsInput struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_apps",
		Description: "List all installed apps managed by TrueNAS SCALE.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listAppsInput) (*mcp.CallToolResult, any, error) {
		apps, err := client.ListApps(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_apps: %w", err)
		}
		return jsonResult(apps)
	})

	type getAppInput struct {
		Name string `json:"name" jsonschema:"required,App name (as shown in the TrueNAS UI)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_app",
		Description: "Get detailed information about a specific app by name.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p getAppInput) (*mcp.CallToolResult, any, error) {
		app, err := client.GetApp(ctx, p.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("get_app: %w", err)
		}
		return jsonResult(app)
	})

	type startAppInput struct {
		Name string `json:"name" jsonschema:"required,App name (as shown in the TrueNAS UI)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "start_app",
		Description: "Start an app by name. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p startAppInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.StartApp(ctx, p.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("start_app: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type stopAppInput struct {
		Name string `json:"name" jsonschema:"required,App name (as shown in the TrueNAS UI)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "stop_app",
		Description: "Stop a running app by name. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p stopAppInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.StopApp(ctx, p.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("stop_app: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type restartAppInput struct {
		Name string `json:"name" jsonschema:"required,App name (as shown in the TrueNAS UI)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "restart_app",
		Description: "Restart an app by name (redeploy). Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p restartAppInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.RestartApp(ctx, p.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("restart_app: %w", err)
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

	type installAppInput struct {
		AppName    string `json:"app_name"          jsonschema:"required,Instance name for the new app (lowercase, hyphens allowed, max 40 chars)"`
		CatalogApp string `json:"catalog_app"       jsonschema:"required,Catalog app to install (e.g. jellyfin)"`
		Train      string `json:"train,omitempty"   jsonschema:"Catalog train (default: stable)"`
		Version    string `json:"version,omitempty" jsonschema:"App version to install (default: latest)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "install_app",
		Description: "Install a catalog app from the TrueNAS app catalog. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p installAppInput) (*mcp.CallToolResult, any, error) {
		train := p.Train
		if train == "" {
			train = "stable"
		}
		version := p.Version
		if version == "" {
			version = "latest"
		}
		jobID, err := client.CreateApp(ctx, &truenas.CreateAppParams{
			AppName:    p.AppName,
			CatalogApp: p.CatalogApp,
			Train:      train,
			Version:    version,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("install_app: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type installCustomAppInput struct {
		AppName                   string `json:"app_name"                     jsonschema:"required,Instance name for the new app (lowercase, hyphens allowed, max 40 chars)"`
		CustomComposeConfigString string `json:"custom_compose_config_string"  jsonschema:"required,Docker Compose YAML defining the services to deploy"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "install_custom_app",
		Description: "Install a custom Docker Compose app on TrueNAS SCALE. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p installCustomAppInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.CreateApp(ctx, &truenas.CreateAppParams{
			AppName:                   p.AppName,
			CustomApp:                 true,
			CustomComposeConfigString: p.CustomComposeConfigString,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("install_custom_app: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type upgradeAppInput struct {
		Name    string `json:"name"              jsonschema:"required,App name (as shown in the TrueNAS UI)"`
		Version string `json:"version,omitempty" jsonschema:"Target version to upgrade to (defaults to latest available)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "upgrade_app",
		Description: "Upgrade an installed app to the specified version, or to the latest available version if version is omitted. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p upgradeAppInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.UpgradeApp(ctx, p.Name, p.Version)
		if err != nil {
			return nil, nil, fmt.Errorf("upgrade_app: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})

	type upgradeSummaryInput struct {
		Name string `json:"name" jsonschema:"required,App name (as shown in the TrueNAS UI)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "upgrade_summary",
		Description: "Get upgrade availability and changelog for an installed app.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p upgradeSummaryInput) (*mcp.CallToolResult, any, error) {
		summary, err := client.GetUpgradeSummary(ctx, p.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("upgrade_summary: %w", err)
		}
		return jsonResult(summary)
	})

	type rollbackAppInput struct {
		Name    string `json:"name"    jsonschema:"required,App name (as shown in the TrueNAS UI)"`
		Version string `json:"version" jsonschema:"required,Version to roll back to"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rollback_app",
		Description: "Roll an app back to a previous version. Returns the async job ID immediately (non-blocking).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: false},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p rollbackAppInput) (*mcp.CallToolResult, any, error) {
		jobID, err := client.RollbackApp(ctx, p.Name, p.Version)
		if err != nil {
			return nil, nil, fmt.Errorf("rollback_app: %w", err)
		}
		return jsonResult(map[string]int{"job_id": jobID})
	})
}
