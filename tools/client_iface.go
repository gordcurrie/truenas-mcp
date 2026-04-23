package tools

import (
	"context"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

// truenasClient is the subset of *truenas.Client used by the tools layer.
// It exists to allow mock implementations in tests without requiring a live TrueNAS server.
type truenasClient interface {
	// System
	GetSystemInfo(ctx context.Context) (*truenas.SystemInfo, error)

	// Pools
	ListPools(ctx context.Context, opts ...truenas.ListOptions) ([]truenas.Pool, error)
	GetPool(ctx context.Context, id int) (*truenas.Pool, error)

	// Datasets
	ListDatasets(ctx context.Context, pool string, opts ...truenas.ListOptions) ([]truenas.Dataset, error)
	GetDataset(ctx context.Context, id string) (*truenas.Dataset, error)
	CreateDataset(ctx context.Context, params *truenas.CreateDatasetParams) (*truenas.Dataset, error)

	// Snapshots
	ListSnapshots(ctx context.Context, dataset string, opts ...truenas.ListOptions) ([]truenas.Snapshot, error)
	GetSnapshot(ctx context.Context, id string) (*truenas.Snapshot, error)
	CreateSnapshot(ctx context.Context, params truenas.CreateSnapshotParams) (*truenas.Snapshot, error)
	RollbackSnapshot(ctx context.Context, id string, params truenas.RollbackSnapshotParams) error
	DeleteSnapshot(ctx context.Context, id string) error

	// Virtual Machines
	ListVMs(ctx context.Context, opts ...truenas.ListOptions) ([]truenas.VM, error)
	GetVM(ctx context.Context, id int) (*truenas.VM, error)
	StartVM(ctx context.Context, id int) (int, error)
	StopVM(ctx context.Context, id int, force bool) (int, error)
	RestartVM(ctx context.Context, id int) (int, error)
	CreateVM(ctx context.Context, params *truenas.CreateVMParams) (*truenas.VM, error)
	UpdateVM(ctx context.Context, id int, params *truenas.UpdateVMParams) (*truenas.VM, error)
	DeleteVM(ctx context.Context, id int) error
	ListVMDevices(ctx context.Context, vmID int) ([]truenas.VMDevice, error)
	AddVMDevice(ctx context.Context, params *truenas.AddVMDeviceParams) (*truenas.VMDevice, error)
	DeleteVMDevice(ctx context.Context, deviceID int) error

	// Apps
	ListApps(ctx context.Context, opts ...truenas.ListOptions) ([]truenas.App, error)
	GetApp(ctx context.Context, name string) (*truenas.App, error)
	StartApp(ctx context.Context, name string) (int, error)
	StopApp(ctx context.Context, name string) (int, error)
	RestartApp(ctx context.Context, name string) (int, error)
	ListImages(ctx context.Context) ([]truenas.Image, error)
	CreateApp(ctx context.Context, params *truenas.CreateAppParams) (int, error)
	DeleteApp(ctx context.Context, name string) error
	UpgradeApp(ctx context.Context, name, version string) (int, error)
	GetUpgradeSummary(ctx context.Context, name string) (*truenas.AppUpgradeSummary, error)
	RollbackApp(ctx context.Context, name, version string) (int, error)

	// Network
	ListInterfaces(ctx context.Context) ([]truenas.Interface, error)

	// Filesystem
	ListDirectory(ctx context.Context, path string) ([]truenas.DirEntry, error)
}
