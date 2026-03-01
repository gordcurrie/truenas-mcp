package truenas

import (
	"context"
	"fmt"
	"regexp"
)

// vmNameRe matches valid TrueNAS VM names: alphanumeric characters only.
var vmNameRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

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

// CreateVMParams holds the fields for creating a new VM via POST /vm.
// Name and Memory are required; all other fields use TrueNAS defaults when omitted.
type CreateVMParams struct {
	// Name is required.
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// VCPUs is the total virtual CPU count; defaults to 1 when omitted.
	VCPUs int `json:"vcpus,omitempty"`
	// Memory is required; specified in MiB.
	Memory          int    `json:"memory"`
	Bootloader      string `json:"bootloader,omitempty"`       // UEFI (default), UEFI_CSM, GRUB
	Autostart       bool   `json:"autostart,omitempty"`        // start with system
	Cores           int    `json:"cores,omitempty"`            // cores per socket
	Threads         int    `json:"threads,omitempty"`          // threads per core
	ShutdownTimeout int    `json:"shutdown_timeout,omitempty"` // seconds before force kill
	CPUMode         string `json:"cpu_mode,omitempty"`         // CUSTOM, HOST-MODEL, HOST-PASSTHROUGH
	CPUModel        string `json:"cpu_model,omitempty"`        // only used when CPUMode=CUSTOM
}

// UpdateVMParams holds the fields that can be updated on an existing VM via
// PUT /vm/id/{id}. Only non-zero/non-empty fields are serialised and sent.
type UpdateVMParams struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	VCPUs           int    `json:"vcpus,omitempty"`
	Memory          int    `json:"memory,omitempty"` // MiB
	Bootloader      string `json:"bootloader,omitempty"`
	Cores           int    `json:"cores,omitempty"`
	Threads         int    `json:"threads,omitempty"`
	ShutdownTimeout int    `json:"shutdown_timeout,omitempty"`
	CPUMode         string `json:"cpu_mode,omitempty"`
	CPUModel        string `json:"cpu_model,omitempty"`
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

// CreateVM creates a new virtual machine and returns it.
// params.Name and params.Memory are required.
func (c *Client) CreateVM(ctx context.Context, params *CreateVMParams) (*VM, error) {
	if params == nil {
		return nil, fmt.Errorf("creating VM: params must not be nil")
	}
	if params.Name == "" {
		return nil, fmt.Errorf("creating VM: name must not be empty")
	}
	if !vmNameRe.MatchString(params.Name) {
		return nil, fmt.Errorf("creating VM: name must contain only alphanumeric characters")
	}
	if params.Memory <= 0 {
		return nil, fmt.Errorf("creating VM: memory must be greater than 0")
	}

	var vm VM
	if err := c.postWithBody(ctx, "/vm", params, &vm); err != nil {
		return nil, fmt.Errorf("creating VM %q: %w", params.Name, err)
	}
	return &vm, nil
}

// UpdateVM updates an existing VM by ID and returns the updated VM.
// Only fields set in params are sent; omitted fields are unchanged.
func (c *Client) UpdateVM(ctx context.Context, id int, params *UpdateVMParams) (*VM, error) {
	if params == nil {
		return nil, fmt.Errorf("updating VM %d: params must not be nil", id)
	}
	if params.Name != "" && !vmNameRe.MatchString(params.Name) {
		return nil, fmt.Errorf("updating VM %d: name must contain only alphanumeric characters", id)
	}
	if params.Memory < 0 {
		return nil, fmt.Errorf("updating VM %d: memory must be greater than 0 when provided", id)
	}
	var vm VM
	if err := c.put(ctx, fmt.Sprintf("/vm/id/%d", id), params, &vm); err != nil {
		return nil, fmt.Errorf("updating VM %d: %w", id, err)
	}
	return &vm, nil
}

// DeleteVM permanently deletes the VM with the given ID.
// The VM must be stopped before deletion.
func (c *Client) DeleteVM(ctx context.Context, id int) error {
	if err := c.delete(ctx, fmt.Sprintf("/vm/id/%d", id), nil); err != nil {
		return fmt.Errorf("deleting VM %d: %w", id, err)
	}
	return nil
}
