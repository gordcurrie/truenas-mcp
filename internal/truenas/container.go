package truenas

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
)

// appNameRe is the validation pattern for TrueNAS app names.
// Must start with a lowercase letter, end with alphanumeric, hyphens allowed in the middle.
// Maximum length of 40 characters is enforced separately.
var appNameRe = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)

// Container represents a TrueNAS SCALE app (Docker-based container).
// Apps are identified by name, not numeric ID.
type Container struct {
	Name  string `json:"name"`
	State string `json:"state"` // RUNNING, STOPPED, DEPLOYING, CRASHED, etc.
}

// Image represents a Docker image stored on TrueNAS SCALE.
type Image struct {
	ID       string   `json:"id"`
	RepoTags []string `json:"repo_tags"`
	Size     int64    `json:"size"`
}

// ListContainers returns all Docker apps (containers) on the TrueNAS SCALE system.
func (c *Client) ListContainers(ctx context.Context) ([]Container, error) {
	var containers []Container
	if err := c.get(ctx, "/app", &containers); err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}
	return containers, nil
}

// GetContainer returns a single app (container) by name.
func (c *Client) GetContainer(ctx context.Context, name string) (*Container, error) {
	var container Container
	if err := c.get(ctx, "/app/id/"+url.PathEscape(name), &container); err != nil {
		return nil, fmt.Errorf("getting container %q: %w", name, err)
	}
	return &container, nil
}

// StartContainer starts the app with the given name and returns the async job ID.
// The job can be polled for completion using PollJob.
func (c *Client) StartContainer(ctx context.Context, name string) (int, error) {
	var jobID int
	if err := c.post(ctx, "/app/id/"+url.PathEscape(name)+"/start", &jobID); err != nil {
		return 0, fmt.Errorf("starting container %q: %w", name, err)
	}
	return jobID, nil
}

// StopContainer stops the app with the given name and returns the async job ID.
// The job can be polled for completion using PollJob.
func (c *Client) StopContainer(ctx context.Context, name string) (int, error) {
	var jobID int
	if err := c.post(ctx, "/app/id/"+url.PathEscape(name)+"/stop", &jobID); err != nil {
		return 0, fmt.Errorf("stopping container %q: %w", name, err)
	}
	return jobID, nil
}

// RestartContainer restarts the app with the given name and returns the async job ID.
// The job can be polled for completion using PollJob.
func (c *Client) RestartContainer(ctx context.Context, name string) (int, error) {
	var jobID int
	if err := c.post(ctx, "/app/id/"+url.PathEscape(name)+"/restart", &jobID); err != nil {
		return 0, fmt.Errorf("restarting container %q: %w", name, err)
	}
	return jobID, nil
}

// ListImages returns all Docker images stored on the TrueNAS SCALE system.
func (c *Client) ListImages(ctx context.Context) ([]Image, error) {
	var images []Image
	if err := c.get(ctx, "/app/image", &images); err != nil {
		return nil, fmt.Errorf("listing images: %w", err)
	}
	return images, nil
}

// CreateContainerParams holds the fields for installing a TrueNAS app via POST /app.
// Set CustomApp=false for catalog apps (CatalogApp required).
// Set CustomApp=true and provide CustomComposeConfigString for custom Docker Compose apps.
type CreateContainerParams struct {
	// AppName is required. Must match ^[a-z]([-a-z0-9]*[a-z0-9])?$ (max 40 chars).
	AppName string `json:"app_name"`
	// CatalogApp is the catalog app name to install. Required when CustomApp is false.
	CatalogApp string `json:"catalog_app,omitempty"`
	// Train is the catalog train (e.g. "stable", "community"). Defaults to "stable".
	Train string `json:"train,omitempty"`
	// Version is the app version to install. Defaults to "latest".
	Version string `json:"version,omitempty"`
	// CustomApp when true installs a custom Docker Compose app instead of a catalog app.
	CustomApp bool `json:"custom_app,omitempty"`
	// CustomComposeConfigString is the raw Docker Compose YAML. Required when CustomApp is true.
	CustomComposeConfigString string `json:"custom_compose_config_string,omitempty"`
	// Values are optional app-specific configuration values for catalog apps.
	Values map[string]any `json:"values,omitempty"`
}

// CreateContainer installs a new TrueNAS app and returns the created container once the
// job is accepted. The returned job ID can be polled for completion using PollJob.
func (c *Client) CreateContainer(ctx context.Context, params *CreateContainerParams) (int, error) {
	if params == nil {
		return 0, fmt.Errorf("creating container: params must not be nil")
	}
	if params.AppName == "" {
		return 0, fmt.Errorf("creating container: app_name must not be empty")
	}
	if len(params.AppName) > 40 {
		return 0, fmt.Errorf("creating container: app_name must not exceed 40 characters")
	}
	if !appNameRe.MatchString(params.AppName) {
		return 0, fmt.Errorf("creating container: app_name must match ^[a-z]([-a-z0-9]*[a-z0-9])?$")
	}
	if !params.CustomApp && params.CatalogApp == "" {
		return 0, fmt.Errorf("creating container: catalog_app is required when custom_app is false")
	}
	if params.CustomApp && params.CustomComposeConfigString == "" {
		return 0, fmt.Errorf("creating container: custom_compose_config_string is required when custom_app is true")
	}

	var jobID int
	if err := c.postWithBody(ctx, "/app", params, &jobID); err != nil {
		return 0, fmt.Errorf("creating container %q: %w", params.AppName, err)
	}
	return jobID, nil
}

// DeleteContainer permanently removes the named app from TrueNAS SCALE.
func (c *Client) DeleteContainer(ctx context.Context, name string) error {
	if err := c.delete(ctx, "/app/id/"+url.PathEscape(name), nil); err != nil {
		return fmt.Errorf("deleting container %q: %w", name, err)
	}
	return nil
}
