package truenas

import (
	"context"
	"fmt"
	"maps"
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
// Pass a ListOptions value to apply server-side pagination (limit / offset).
func (c *Client) ListVMs(ctx context.Context, opts ...ListOptions) ([]VM, error) {
	var o ListOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	qs, err := buildQueryString(nil, o)
	if err != nil {
		return nil, fmt.Errorf("listing VMs: %w", err)
	}
	var vms []VM
	if err := c.get(ctx, "/vm"+qs, &vms); err != nil {
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
	// Memory uses omitempty, so 0 means "leave unchanged" and is never serialised.
	// Only explicitly negative values are rejected — they would be nonsensical and
	// are distinct from 0 (which an LLM would send if it wants to clear the field,
	// which is unsupported; the tool description notes this limitation).
	if params.Memory < 0 {
		return nil, fmt.Errorf("updating VM %d: memory must be positive when provided", id)
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
	if err := c.delete(ctx, fmt.Sprintf("/vm/id/%d", id)); err != nil {
		return fmt.Errorf("deleting VM %d: %w", id, err)
	}
	return nil
}

// VMDeviceType identifies the kind of hardware device attached to a VM.
type VMDeviceType string

const (
	// VMDeviceTypeDISK is a virtual disk backed by a ZFS zvol.
	VMDeviceTypeDISK VMDeviceType = "DISK"
	// VMDeviceTypeCDROM is a virtual optical drive backed by an ISO file on disk.
	VMDeviceTypeCDROM VMDeviceType = "CDROM"
	// VMDeviceTypeNIC is a virtual network interface card.
	VMDeviceTypeNIC VMDeviceType = "NIC"
	// VMDeviceTypeDISPLAY is a virtual display adapter (VNC/SPICE).
	VMDeviceTypeDISPLAY VMDeviceType = "DISPLAY"
	// VMDeviceTypeRAW is a raw block device passthrough.
	VMDeviceTypeRAW VMDeviceType = "RAW"
)

// VMDevice represents a hardware device attached to a TrueNAS SCALE VM.
//
// Attributes are device-type-specific:
//
//	DISK:    {"path": "zvol/pool/dataset", "type": "VIRTIO"}
//	CDROM:   {"path": "/mnt/Storage/isos/debian.iso"}
//	NIC:     {"type": "VIRTIO", "nic_attach": "br0"}
//	DISPLAY: {"web": true, "port": 5900, "web_port": 5901}
type VMDevice struct {
	ID         int            `json:"id"`
	VMID       int            `json:"vm"`
	DType      VMDeviceType   `json:"dtype"`
	Attributes map[string]any `json:"attributes"`
	Order      *int           `json:"order"`
}

// AddVMDeviceParams holds the parameters for attaching a new device to a VM.
// VMID, DType, and Attributes are required.
//
// Attributes by device type:
//
//	DISK:    {"path": "zvol/pool/dataset", "type": "VIRTIO"}
//	CDROM:   {"path": "/mnt/Storage/isos/pbs.iso"}
//	NIC:     {"type": "VIRTIO", "nic_attach": "br0"}
//	DISPLAY: {"web": true, "port": 5900}
type AddVMDeviceParams struct {
	VMID       int            `json:"vm"`
	DType      VMDeviceType   `json:"dtype"`
	Attributes map[string]any `json:"attributes"`
	Order      *int           `json:"order,omitempty"`
}

// addVMDeviceWire is the JSON shape the TrueNAS API actually expects:
// dtype lives inside attributes, not at the top level.
type addVMDeviceWire struct {
	VMID       int            `json:"vm"`
	Attributes map[string]any `json:"attributes"`
	Order      *int           `json:"order,omitempty"`
}

// ListVMDevices returns all devices attached to the VM with the given ID.
// It fetches all devices from the API and filters them by VM ID client-side.
func (c *Client) ListVMDevices(ctx context.Context, vmID int) ([]VMDevice, error) {
	var all []VMDevice
	if err := c.get(ctx, "/vm/device", &all); err != nil {
		return nil, fmt.Errorf("listing VM devices: %w", err)
	}

	var devices []VMDevice
	for _, d := range all {
		if d.VMID == vmID {
			devices = append(devices, d)
		}
	}
	return devices, nil
}

// AddVMDevice attaches a new device to a VM and returns the created VMDevice.
// params.VMID, params.DType, and params.Attributes are required.
func (c *Client) AddVMDevice(ctx context.Context, params *AddVMDeviceParams) (*VMDevice, error) {
	if params == nil {
		return nil, fmt.Errorf("adding VM device: params must not be nil")
	}
	if params.VMID <= 0 {
		return nil, fmt.Errorf("adding VM device: vm must be a positive integer")
	}
	if params.DType == "" {
		return nil, fmt.Errorf("adding VM device: dtype must not be empty")
	}
	if len(params.Attributes) == 0 {
		return nil, fmt.Errorf("adding VM device: attributes must not be empty")
	}

	// The TrueNAS API expects dtype inside the attributes map, not as a
	// top-level field. Build the wire payload accordingly.
	// Reject ambiguous input: DType is the authoritative source for dtype.
	// Callers must not set "dtype" inside Attributes directly.
	if _, conflict := params.Attributes["dtype"]; conflict {
		return nil, fmt.Errorf("adding VM device: dtype must be set via DType, not inside Attributes")
	}
	attrs := make(map[string]any, len(params.Attributes)+1)
	maps.Copy(attrs, params.Attributes)
	attrs["dtype"] = string(params.DType)

	wire := addVMDeviceWire{
		VMID:       params.VMID,
		Attributes: attrs,
		Order:      params.Order,
	}

	var device VMDevice
	if err := c.postWithBody(ctx, "/vm/device", wire, &device); err != nil {
		return nil, fmt.Errorf("adding %s device to VM %d: %w", params.DType, params.VMID, err)
	}
	return &device, nil
}

// DeleteVMDevice removes the device with the given ID from its VM.
func (c *Client) DeleteVMDevice(ctx context.Context, deviceID int) error {
	if err := c.delete(ctx, fmt.Sprintf("/vm/device/id/%d", deviceID)); err != nil {
		return fmt.Errorf("deleting VM device %d: %w", deviceID, err)
	}
	return nil
}
