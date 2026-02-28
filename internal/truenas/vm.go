package truenas

import (
	"context"
	"fmt"
)

// VMStatus holds the runtime state of a VM as reported by TrueNAS SCALE.
type VMStatus struct {
	State       string `json:"state"`        // RUNNING, STOPPED
	PID         *int   `json:"pid"`          // nil when not running
	DomainState string `json:"domain_state"` // libvirt domain state string
}

// VM represents a virtual machine configured on TrueNAS SCALE.
type VM struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	VCPUs           int      `json:"vcpus"`
	Memory          int      `json:"memory"` // MiB
	Autostart       bool     `json:"autostart"`
	Bootloader      string   `json:"bootloader"`
	Cores           int      `json:"cores"`
	Threads         int      `json:"threads"`
	ShutdownTimeout int      `json:"shutdown_timeout"`
	CPUMode         string   `json:"cpu_mode"`
	CPUModel        string   `json:"cpu_model"`
	UUID            string   `json:"uuid"`
	Status          VMStatus `json:"status"`
}

// stopBody is the request body for the stop VM endpoint.
type stopBody struct {
	Force bool `json:"force"`
}

// ListVMs returns all virtual machines configured on the TrueNAS SCALE system.
func (c *Client) ListVMs(ctx context.Context) ([]VM, error) {
	var vms []VM
	if err := c.get(ctx, "/vm", &vms); err != nil {
		return nil, fmt.Errorf("listing VMs: %w", err)
	}
	return vms, nil
}

// GetVM returns a single VM by its numeric ID.
func (c *Client) GetVM(ctx context.Context, id int) (*VM, error) {
	var vm VM
	if err := c.get(ctx, fmt.Sprintf("/vm/id/%d", id), &vm); err != nil {
		return nil, fmt.Errorf("getting VM %d: %w", id, err)
	}
	return &vm, nil
}

// StartVM starts the VM with the given ID and returns the async job ID.
// The job can be polled for completion using PollJob.
func (c *Client) StartVM(ctx context.Context, id int) (int, error) {
	var jobID int
	if err := c.post(ctx, fmt.Sprintf("/vm/id/%d/start", id), &jobID); err != nil {
		return 0, fmt.Errorf("starting VM %d: %w", id, err)
	}
	return jobID, nil
}

// StopVM stops the VM with the given ID and returns the async job ID.
// Set force to true to forcibly terminate without a graceful shutdown.
// The job can be polled for completion using PollJob.
func (c *Client) StopVM(ctx context.Context, id int, force bool) (int, error) {
	var jobID int
	if err := c.postWithBody(ctx, fmt.Sprintf("/vm/id/%d/stop", id), stopBody{Force: force}, &jobID); err != nil {
		return 0, fmt.Errorf("stopping VM %d: %w", id, err)
	}
	return jobID, nil
}

// RestartVM restarts the VM with the given ID and returns the async job ID.
// The job can be polled for completion using PollJob.
func (c *Client) RestartVM(ctx context.Context, id int) (int, error) {
	var jobID int
	if err := c.post(ctx, fmt.Sprintf("/vm/id/%d/restart", id), &jobID); err != nil {
		return 0, fmt.Errorf("restarting VM %d: %w", id, err)
	}
	return jobID, nil
}
