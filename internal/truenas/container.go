package truenas

import (
	"context"
	"fmt"
)

// ContainerStatus holds the runtime state of a Docker container.
type ContainerStatus struct {
	State string `json:"state"` // RUNNING, STOPPED, etc.
}

// Container represents a Docker container managed by TrueNAS SCALE.
type Container struct {
	ID     int             `json:"id"`
	Name   string          `json:"name"`
	Image  string          `json:"image"`
	Status ContainerStatus `json:"status"`
	// AutoStart controls whether the container starts with the system.
	AutoStart bool `json:"autostart"`
}

// Image represents a Docker image stored on TrueNAS SCALE.
type Image struct {
	ID       int      `json:"id"`
	RepoTags []string `json:"repo_tags"`
	Size     int64    `json:"size"`
}

// ListContainers returns all Docker containers configured on the TrueNAS SCALE system.
func (c *Client) ListContainers(ctx context.Context) ([]Container, error) {
	var containers []Container
	if err := c.get(ctx, "/container/container", &containers); err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}
	return containers, nil
}

// GetContainer returns a single container by its numeric ID.
func (c *Client) GetContainer(ctx context.Context, id int) (*Container, error) {
	var container Container
	if err := c.get(ctx, fmt.Sprintf("/container/container/id/%d", id), &container); err != nil {
		return nil, fmt.Errorf("getting container %d: %w", id, err)
	}
	return &container, nil
}

// StartContainer starts the container with the given ID and returns the async job ID.
// The job can be polled for completion using PollJob.
func (c *Client) StartContainer(ctx context.Context, id int) (int, error) {
	var jobID int
	if err := c.post(ctx, fmt.Sprintf("/container/container/id/%d/start", id), &jobID); err != nil {
		return 0, fmt.Errorf("starting container %d: %w", id, err)
	}
	return jobID, nil
}

// StopContainer stops the container with the given ID and returns the async job ID.
// The job can be polled for completion using PollJob.
func (c *Client) StopContainer(ctx context.Context, id int) (int, error) {
	var jobID int
	if err := c.post(ctx, fmt.Sprintf("/container/container/id/%d/stop", id), &jobID); err != nil {
		return 0, fmt.Errorf("stopping container %d: %w", id, err)
	}
	return jobID, nil
}

// RestartContainer restarts the container with the given ID and returns the async job ID.
// The job can be polled for completion using PollJob.
func (c *Client) RestartContainer(ctx context.Context, id int) (int, error) {
	var jobID int
	if err := c.post(ctx, fmt.Sprintf("/container/container/id/%d/restart", id), &jobID); err != nil {
		return 0, fmt.Errorf("restarting container %d: %w", id, err)
	}
	return jobID, nil
}

// ListImages returns all Docker images stored on the TrueNAS SCALE system.
func (c *Client) ListImages(ctx context.Context) ([]Image, error) {
	var images []Image
	if err := c.get(ctx, "/container/image", &images); err != nil {
		return nil, fmt.Errorf("listing images: %w", err)
	}
	return images, nil
}
