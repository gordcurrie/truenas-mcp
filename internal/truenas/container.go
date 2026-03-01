package truenas

import (
	"context"
	"fmt"
	"net/url"
)

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
