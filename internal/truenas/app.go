package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
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
type App struct {
	Name  string `json:"name"`
	State string `json:"state"` // RUNNING, STOPPED, DEPLOYING, CRASHED, etc.
}

// Image represents a Docker image stored on TrueNAS SCALE.
type Image struct {
	ID       string   `json:"id"`
	RepoTags []string `json:"repo_tags,omitempty"`
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
// UpgradeAvailable is false when the app is already on the latest version.
type AppUpgradeSummary struct {
	UpgradeAvailable            bool             `json:"upgrade_available"`
	LatestVersion               string           `json:"latest_version,omitempty"`
	LatestHumanVersion          string           `json:"latest_human_version,omitempty"`
	UpgradeVersion              string           `json:"upgrade_version,omitempty"`
	UpgradeHumanVersion         string           `json:"upgrade_human_version,omitempty"`
	AvailableVersionsForUpgrade []AppVersionInfo `json:"available_versions_for_upgrade,omitempty"`
	Changelog                   *string          `json:"changelog,omitempty"`
}

// ListApps returns all installed apps on the TrueNAS SCALE system.
// Pass a ListOptions value to apply server-side pagination (limit / offset).
func (c *Client) ListApps(ctx context.Context, opts ...ListOptions) ([]App, error) {
	if len(opts) > 1 {
		return nil, fmt.Errorf("listing apps: at most one ListOptions value may be provided")
	}
	var o ListOptions
	if len(opts) == 1 {
		o = opts[0]
	}
	if err := validateListOptions(o); err != nil {
		return nil, fmt.Errorf("listing apps: %w", err)
	}
	var apps []App
	if err := c.call(ctx, "app.query", buildQueryParams(nil, o), &apps); err != nil {
		return nil, fmt.Errorf("listing apps: %w", err)
	}
	return apps, nil
}

// GetApp returns a single app by name.
func (c *Client) GetApp(ctx context.Context, name string) (*App, error) {
	var app App
	if err := c.call(ctx, "app.get_instance", []any{name}, &app); err != nil {
		return nil, fmt.Errorf("getting app %q: %w", name, err)
	}
	return &app, nil
}

// StartApp starts the app with the given name and returns the job result.
// See callJob for the shape of the result.
func (c *Client) StartApp(ctx context.Context, name string) (json.RawMessage, error) {
	res, err := c.callJob(ctx, "app.start", []any{name})
	if err != nil {
		return nil, fmt.Errorf("starting app %q: %w", name, err)
	}
	return res, nil
}

// StopApp stops the app with the given name and returns the job result.
// See callJob for the shape of the result.
func (c *Client) StopApp(ctx context.Context, name string) (json.RawMessage, error) {
	res, err := c.callJob(ctx, "app.stop", []any{name})
	if err != nil {
		return nil, fmt.Errorf("stopping app %q: %w", name, err)
	}
	return res, nil
}

// RestartApp redeploys (restarts) the app with the given name and returns the job result.
// TrueNAS SCALE uses app.redeploy (not app.restart) for this operation.
// See callJob for the shape of the result.
func (c *Client) RestartApp(ctx context.Context, name string) (json.RawMessage, error) {
	res, err := c.callJob(ctx, "app.redeploy", []any{name})
	if err != nil {
		return nil, fmt.Errorf("restarting app %q: %w", name, err)
	}
	return res, nil
}

// ListImages returns all Docker images stored on the TrueNAS SCALE system.
func (c *Client) ListImages(ctx context.Context) ([]Image, error) {
	var images []Image
	if err := c.call(ctx, "app.image.query", buildQueryParams(nil, ListOptions{}), &images); err != nil {
		return nil, fmt.Errorf("listing images: %w", err)
	}
	return images, nil
}

// CreateAppParams holds the fields for installing a TrueNAS app.
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

// CreateApp starts installing a new TrueNAS app and returns the job result.
// See callJob for the shape of the result.
func (c *Client) CreateApp(ctx context.Context, params *CreateAppParams) (json.RawMessage, error) {
	if params == nil {
		return nil, fmt.Errorf("creating app: params must not be nil")
	}
	if err := validateAppName(params.AppName, "app_name"); err != nil {
		return nil, err
	}
	if !params.CustomApp && params.CatalogApp == "" {
		return nil, fmt.Errorf("creating app: catalog_app is required when custom_app is false")
	}
	if params.CustomApp && params.CustomComposeConfigString == "" {
		return nil, fmt.Errorf("creating app: custom_compose_config_string is required when custom_app is true")
	}

	res, err := c.callJob(ctx, "app.create", []any{params})
	if err != nil {
		return nil, fmt.Errorf("creating app %q: %w", params.AppName, err)
	}
	return res, nil
}

// DeleteApp permanently removes the named app from TrueNAS SCALE.
func (c *Client) DeleteApp(ctx context.Context, name string) error {
	if err := validateAppName(name, "name"); err != nil {
		return err
	}
	if err := c.call(ctx, "app.delete", []any{name}, nil); err != nil {
		return fmt.Errorf("deleting app %q: %w", name, err)
	}
	return nil
}

// UpgradeApp upgrades the named app to the given version, or to the latest available
// version when version is empty. Returns the job result (see callJob).
func (c *Client) UpgradeApp(ctx context.Context, name, version string) (json.RawMessage, error) {
	if err := validateAppName(name, "name"); err != nil {
		return nil, err
	}
	res, err := c.callJob(ctx, "app.upgrade", []any{name, &AppUpgradeParams{AppVersion: version}})
	if err != nil {
		return nil, fmt.Errorf("upgrading app %q: %w", name, err)
	}
	return res, nil
}

// GetUpgradeSummary retrieves the upgrade summary for the named app.
// When the app is already up to date, UpgradeAvailable is false and no error is returned.
func (c *Client) GetUpgradeSummary(ctx context.Context, name string) (*AppUpgradeSummary, error) {
	if err := validateAppName(name, "name"); err != nil {
		return nil, err
	}
	var summary AppUpgradeSummary
	err := c.call(ctx, "app.upgrade_summary", []any{name}, &summary)
	if err != nil {
		if noUpgradeAvailable(err) {
			return &AppUpgradeSummary{UpgradeAvailable: false}, nil
		}
		return nil, fmt.Errorf("getting upgrade summary for app %q: %w", name, err)
	}
	summary.UpgradeAvailable = true
	return &summary, nil
}

// RollbackApp rolls the named app back to the specified previous version.
// Returns the job result (see callJob).
func (c *Client) RollbackApp(ctx context.Context, name, version string) (json.RawMessage, error) {
	if err := validateAppName(name, "name"); err != nil {
		return nil, err
	}
	if version == "" {
		return nil, fmt.Errorf("rolling back app: version must not be empty")
	}
	res, err := c.callJob(ctx, "app.rollback", []any{name, &AppUpgradeParams{AppVersion: version}})
	if err != nil {
		return nil, fmt.Errorf("rolling back app %q: %w", name, err)
	}
	return res, nil
}

// noUpgradeAvailable reports whether err indicates TrueNAS found no pending
// upgrade for the app — a normal condition, not a failure.
func noUpgradeAvailable(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	lower := strings.ToLower(apiErr.Body)
	return strings.Contains(lower, "no update") ||
		strings.Contains(lower, "no upgrade") ||
		strings.Contains(lower, "up to date") ||
		strings.Contains(lower, "up-to-date")
}
