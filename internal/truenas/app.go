package truenas

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

// appNameRe is the validation pattern for TrueNAS app names.
// Must start with a lowercase letter, end with alphanumeric, hyphens allowed in the middle.
// Maximum length of 40 characters is enforced separately.
var appNameRe = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)

// validateAppName checks that name is non-empty, ≤40 chars, and matches appNameRe.
// field is the parameter name used in error messages (e.g. "app_name", "name").
func validateAppName(name, field string) error {
	if name == "" {
		return fmt.Errorf("%s: must not be empty", field)
	}
	if len(name) > 40 {
		return fmt.Errorf("%s: must not exceed 40 characters", field)
	}
	if !appNameRe.MatchString(name) {
		return fmt.Errorf("%s: must match ^[a-z]([-a-z0-9]*[a-z0-9])?$", field)
	}
	return nil
}

// App represents a TrueNAS SCALE app (Docker-based).
// Apps are identified by name, not numeric ID.
// Note: the experimental container UI in TrueNAS 25.10 is separate and shares this endpoint;
// a dedicated container API is expected in a future TrueNAS release.
type App struct {
	Name  string `json:"name"`
	State string `json:"state"` // RUNNING, STOPPED, DEPLOYING, CRASHED, etc.
}

// Image represents a Docker image stored on TrueNAS SCALE.
type Image struct {
	ID       string   `json:"id"`
	RepoTags []string `json:"repo_tags"`
	Size     int64    `json:"size"`
}

// AppUpgradeParams holds the optional target version for an app upgrade or rollback.
type AppUpgradeParams struct {
	// AppVersion is the target version. Optional for upgrade (defaults to latest),
	// required for rollback.
	AppVersion string `json:"app_version,omitempty"`
}

// AppVersionInfo holds a version and its human-readable label.
type AppVersionInfo struct {
	Version      string `json:"version"`
	HumanVersion string `json:"human_version"`
}

// AppUpgradeSummary holds the result of app.upgrade_summary.
// UpgradeAvailable is false when the app is already on the latest version
// (the API returns a 422 in that case rather than a summary object).
type AppUpgradeSummary struct {
	UpgradeAvailable            bool             `json:"upgrade_available"`
	LatestVersion               string           `json:"latest_version"`
	LatestHumanVersion          string           `json:"latest_human_version"`
	UpgradeVersion              string           `json:"upgrade_version"`
	UpgradeHumanVersion         string           `json:"upgrade_human_version"`
	AvailableVersionsForUpgrade []AppVersionInfo `json:"available_versions_for_upgrade"`
	Changelog                   *string          `json:"changelog"`
}

// ListApps returns all installed apps on the TrueNAS SCALE system.
func (c *Client) ListApps(ctx context.Context) ([]App, error) {
	var apps []App
	if err := c.get(ctx, "/app", &apps); err != nil {
		return nil, fmt.Errorf("listing apps: %w", err)
	}
	return apps, nil
}

// GetApp returns a single app by name.
func (c *Client) GetApp(ctx context.Context, name string) (*App, error) {
	var app App
	if err := c.get(ctx, "/app/id/"+url.PathEscape(name), &app); err != nil {
		return nil, fmt.Errorf("getting app %q: %w", name, err)
	}
	return &app, nil
}

// StartApp starts the app with the given name and returns the async job ID.
// Uses POST /app/start with the app name as the JSON body.
// The job can be polled for completion using PollJob.
func (c *Client) StartApp(ctx context.Context, name string) (int, error) {
	var jobID int
	if err := c.postWithBody(ctx, "/app/start", name, &jobID); err != nil {
		return 0, fmt.Errorf("starting app %q: %w", name, err)
	}
	return jobID, nil
}

// StopApp stops the app with the given name and returns the async job ID.
// Uses POST /app/stop with the app name as the JSON body.
// The job can be polled for completion using PollJob.
func (c *Client) StopApp(ctx context.Context, name string) (int, error) {
	var jobID int
	if err := c.postWithBody(ctx, "/app/stop", name, &jobID); err != nil {
		return 0, fmt.Errorf("stopping app %q: %w", name, err)
	}
	return jobID, nil
}

// RestartApp redeploys (restarts) the app with the given name and returns the async job ID.
// TrueNAS SCALE uses POST /app/redeploy (not /restart) with the app name as the JSON body.
// The job can be polled for completion using PollJob.
func (c *Client) RestartApp(ctx context.Context, name string) (int, error) {
	var jobID int
	if err := c.postWithBody(ctx, "/app/redeploy", name, &jobID); err != nil {
		return 0, fmt.Errorf("restarting app %q: %w", name, err)
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

// CreateAppParams holds the fields for installing a TrueNAS app via POST /app.
// Set CustomApp=false for catalog apps (CatalogApp required).
// Set CustomApp=true and provide CustomComposeConfigString for custom Docker Compose apps.
type CreateAppParams struct {
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

// CreateApp starts installing a new TrueNAS app and returns the async job ID once the
// job is accepted. The returned job ID can be polled for completion using PollJob.
func (c *Client) CreateApp(ctx context.Context, params *CreateAppParams) (int, error) {
	if params == nil {
		return 0, fmt.Errorf("creating app: params must not be nil")
	}
	if err := validateAppName(params.AppName, "app_name"); err != nil {
		return 0, err
	}
	if !params.CustomApp && params.CatalogApp == "" {
		return 0, fmt.Errorf("creating app: catalog_app is required when custom_app is false")
	}
	if params.CustomApp && params.CustomComposeConfigString == "" {
		return 0, fmt.Errorf("creating app: custom_compose_config_string is required when custom_app is true")
	}

	var jobID int
	if err := c.postWithBody(ctx, "/app", params, &jobID); err != nil {
		return 0, fmt.Errorf("creating app %q: %w", params.AppName, err)
	}
	return jobID, nil
}

// DeleteApp permanently removes the named app from TrueNAS SCALE.
func (c *Client) DeleteApp(ctx context.Context, name string) error {
	if err := validateAppName(name, "name"); err != nil {
		return err
	}
	if err := c.delete(ctx, "/app/id/"+url.PathEscape(name)); err != nil {
		return fmt.Errorf("deleting app %q: %w", name, err)
	}
	return nil
}

// UpgradeApp upgrades the named app to the given version, or to the latest available
// version when version is empty. Returns the async job ID.
func (c *Client) UpgradeApp(ctx context.Context, name, version string) (int, error) {
	if err := validateAppName(name, "name"); err != nil {
		return 0, err
	}
	var jobID int
	if err := c.postWithBody(ctx, "/app/id/"+url.PathEscape(name)+"/upgrade", &AppUpgradeParams{AppVersion: version}, &jobID); err != nil {
		return 0, fmt.Errorf("upgrading app %q: %w", name, err)
	}
	return jobID, nil
}

// GetUpgradeSummary retrieves the upgrade summary for the named app.
// Uses POST /app/upgrade_summary with {"app_name": name}.
// When the app is already up to date the API returns 422 — in that case
// UpgradeAvailable is false and no error is returned.
func (c *Client) GetUpgradeSummary(ctx context.Context, name string) (*AppUpgradeSummary, error) {
	if err := validateAppName(name, "name"); err != nil {
		return nil, err
	}
	body := struct {
		AppName string `json:"app_name"`
	}{AppName: name}
	var summary AppUpgradeSummary
	err := c.postWithBody(ctx, "/app/upgrade_summary", body, &summary)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnprocessableEntity {
			// 422 means no upgrade is available — this is a normal condition.
			return &AppUpgradeSummary{UpgradeAvailable: false}, nil
		}
		return nil, fmt.Errorf("getting upgrade summary for app %q: %w", name, err)
	}
	summary.UpgradeAvailable = true
	return &summary, nil
}

// RollbackApp rolls the named app back to the specified previous version.
// Returns the async job ID.
func (c *Client) RollbackApp(ctx context.Context, name, version string) (int, error) {
	if err := validateAppName(name, "name"); err != nil {
		return 0, err
	}
	if version == "" {
		return 0, fmt.Errorf("rolling back app: version must not be empty")
	}
	var jobID int
	if err := c.postWithBody(ctx, "/app/id/"+url.PathEscape(name)+"/rollback", &AppUpgradeParams{AppVersion: version}, &jobID); err != nil {
		return 0, fmt.Errorf("rolling back app %q: %w", name, err)
	}
	return jobID, nil
}
