package tools

import (
	"context"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
)

// mockTruenasClient is a test double for truenasClient. Each method delegates to
// the corresponding *Fn field if set; otherwise it returns zero values with no error.
type mockTruenasClient struct {
	getSystemInfoFn func(context.Context) (*truenas.SystemInfo, error)

	listPoolsFn func(context.Context, ...truenas.ListOptions) ([]truenas.Pool, error)
	getPoolFn   func(context.Context, int) (*truenas.Pool, error)

	listDatasetsFn  func(context.Context, string, ...truenas.ListOptions) ([]truenas.Dataset, error)
	getDatasetFn    func(context.Context, string) (*truenas.Dataset, error)
	createDatasetFn func(context.Context, *truenas.CreateDatasetParams) (*truenas.Dataset, error)

	listSnapshotsFn  func(context.Context, string, ...truenas.ListOptions) ([]truenas.Snapshot, error)
	getSnapshotFn    func(context.Context, string) (*truenas.Snapshot, error)
	createSnapshotFn func(context.Context, truenas.CreateSnapshotParams) (*truenas.Snapshot, error)
	rollbackSnapshotFn func(context.Context, string, truenas.RollbackSnapshotParams) error
	deleteSnapshotFn func(context.Context, string) error

	listVMsFn      func(context.Context, ...truenas.ListOptions) ([]truenas.VM, error)
	getVMFn        func(context.Context, int) (*truenas.VM, error)
	startVMFn      func(context.Context, int) (int, error)
	stopVMFn       func(context.Context, int, bool) (int, error)
	restartVMFn    func(context.Context, int) (int, error)
	createVMFn     func(context.Context, *truenas.CreateVMParams) (*truenas.VM, error)
	updateVMFn     func(context.Context, int, *truenas.UpdateVMParams) (*truenas.VM, error)
	deleteVMFn     func(context.Context, int) error
	listVMDevicesFn func(context.Context, int) ([]truenas.VMDevice, error)
	addVMDeviceFn  func(context.Context, *truenas.AddVMDeviceParams) (*truenas.VMDevice, error)
	deleteVMDeviceFn func(context.Context, int) error

	listAppsFn       func(context.Context, ...truenas.ListOptions) ([]truenas.App, error)
	getAppFn         func(context.Context, string) (*truenas.App, error)
	startAppFn       func(context.Context, string) (int, error)
	stopAppFn        func(context.Context, string) (int, error)
	restartAppFn     func(context.Context, string) (int, error)
	listImagesFn     func(context.Context) ([]truenas.Image, error)
	createAppFn      func(context.Context, *truenas.CreateAppParams) (int, error)
	deleteAppFn      func(context.Context, string) error
	upgradeAppFn     func(context.Context, string, string) (int, error)
	getUpgradeSummaryFn func(context.Context, string) (*truenas.AppUpgradeSummary, error)
	rollbackAppFn    func(context.Context, string, string) (int, error)

	listInterfacesFn func(context.Context) ([]truenas.Interface, error)
	listDirectoryFn  func(context.Context, string) ([]truenas.DirEntry, error)
}

func (m *mockTruenasClient) GetSystemInfo(ctx context.Context) (*truenas.SystemInfo, error) {
	if m.getSystemInfoFn != nil {
		return m.getSystemInfoFn(ctx)
	}
	return &truenas.SystemInfo{}, nil
}

func (m *mockTruenasClient) ListPools(ctx context.Context, opts ...truenas.ListOptions) ([]truenas.Pool, error) {
	if m.listPoolsFn != nil {
		return m.listPoolsFn(ctx, opts...)
	}
	return nil, nil
}

func (m *mockTruenasClient) GetPool(ctx context.Context, id int) (*truenas.Pool, error) {
	if m.getPoolFn != nil {
		return m.getPoolFn(ctx, id)
	}
	return &truenas.Pool{}, nil
}

func (m *mockTruenasClient) ListDatasets(ctx context.Context, pool string, opts ...truenas.ListOptions) ([]truenas.Dataset, error) {
	if m.listDatasetsFn != nil {
		return m.listDatasetsFn(ctx, pool, opts...)
	}
	return nil, nil
}

func (m *mockTruenasClient) GetDataset(ctx context.Context, id string) (*truenas.Dataset, error) {
	if m.getDatasetFn != nil {
		return m.getDatasetFn(ctx, id)
	}
	return &truenas.Dataset{}, nil
}

func (m *mockTruenasClient) CreateDataset(ctx context.Context, params *truenas.CreateDatasetParams) (*truenas.Dataset, error) {
	if m.createDatasetFn != nil {
		return m.createDatasetFn(ctx, params)
	}
	return &truenas.Dataset{}, nil
}

func (m *mockTruenasClient) ListSnapshots(ctx context.Context, dataset string, opts ...truenas.ListOptions) ([]truenas.Snapshot, error) {
	if m.listSnapshotsFn != nil {
		return m.listSnapshotsFn(ctx, dataset, opts...)
	}
	return nil, nil
}

func (m *mockTruenasClient) GetSnapshot(ctx context.Context, id string) (*truenas.Snapshot, error) {
	if m.getSnapshotFn != nil {
		return m.getSnapshotFn(ctx, id)
	}
	return &truenas.Snapshot{}, nil
}

func (m *mockTruenasClient) CreateSnapshot(ctx context.Context, params truenas.CreateSnapshotParams) (*truenas.Snapshot, error) {
	if m.createSnapshotFn != nil {
		return m.createSnapshotFn(ctx, params)
	}
	return &truenas.Snapshot{}, nil
}

func (m *mockTruenasClient) RollbackSnapshot(ctx context.Context, id string, params truenas.RollbackSnapshotParams) error {
	if m.rollbackSnapshotFn != nil {
		return m.rollbackSnapshotFn(ctx, id, params)
	}
	return nil
}

func (m *mockTruenasClient) DeleteSnapshot(ctx context.Context, id string) error {
	if m.deleteSnapshotFn != nil {
		return m.deleteSnapshotFn(ctx, id)
	}
	return nil
}

func (m *mockTruenasClient) ListVMs(ctx context.Context, opts ...truenas.ListOptions) ([]truenas.VM, error) {
	if m.listVMsFn != nil {
		return m.listVMsFn(ctx, opts...)
	}
	return nil, nil
}

func (m *mockTruenasClient) GetVM(ctx context.Context, id int) (*truenas.VM, error) {
	if m.getVMFn != nil {
		return m.getVMFn(ctx, id)
	}
	return &truenas.VM{}, nil
}

func (m *mockTruenasClient) StartVM(ctx context.Context, id int) (int, error) {
	if m.startVMFn != nil {
		return m.startVMFn(ctx, id)
	}
	return 0, nil
}

func (m *mockTruenasClient) StopVM(ctx context.Context, id int, force bool) (int, error) {
	if m.stopVMFn != nil {
		return m.stopVMFn(ctx, id, force)
	}
	return 0, nil
}

func (m *mockTruenasClient) RestartVM(ctx context.Context, id int) (int, error) {
	if m.restartVMFn != nil {
		return m.restartVMFn(ctx, id)
	}
	return 0, nil
}

func (m *mockTruenasClient) CreateVM(ctx context.Context, params *truenas.CreateVMParams) (*truenas.VM, error) {
	if m.createVMFn != nil {
		return m.createVMFn(ctx, params)
	}
	return &truenas.VM{}, nil
}

func (m *mockTruenasClient) UpdateVM(ctx context.Context, id int, params *truenas.UpdateVMParams) (*truenas.VM, error) {
	if m.updateVMFn != nil {
		return m.updateVMFn(ctx, id, params)
	}
	return &truenas.VM{}, nil
}

func (m *mockTruenasClient) DeleteVM(ctx context.Context, id int) error {
	if m.deleteVMFn != nil {
		return m.deleteVMFn(ctx, id)
	}
	return nil
}

func (m *mockTruenasClient) ListVMDevices(ctx context.Context, vmID int) ([]truenas.VMDevice, error) {
	if m.listVMDevicesFn != nil {
		return m.listVMDevicesFn(ctx, vmID)
	}
	return nil, nil
}

func (m *mockTruenasClient) AddVMDevice(ctx context.Context, params *truenas.AddVMDeviceParams) (*truenas.VMDevice, error) {
	if m.addVMDeviceFn != nil {
		return m.addVMDeviceFn(ctx, params)
	}
	return &truenas.VMDevice{}, nil
}

func (m *mockTruenasClient) DeleteVMDevice(ctx context.Context, deviceID int) error {
	if m.deleteVMDeviceFn != nil {
		return m.deleteVMDeviceFn(ctx, deviceID)
	}
	return nil
}

func (m *mockTruenasClient) ListApps(ctx context.Context, opts ...truenas.ListOptions) ([]truenas.App, error) {
	if m.listAppsFn != nil {
		return m.listAppsFn(ctx, opts...)
	}
	return nil, nil
}

func (m *mockTruenasClient) GetApp(ctx context.Context, name string) (*truenas.App, error) {
	if m.getAppFn != nil {
		return m.getAppFn(ctx, name)
	}
	return &truenas.App{}, nil
}

func (m *mockTruenasClient) StartApp(ctx context.Context, name string) (int, error) {
	if m.startAppFn != nil {
		return m.startAppFn(ctx, name)
	}
	return 0, nil
}

func (m *mockTruenasClient) StopApp(ctx context.Context, name string) (int, error) {
	if m.stopAppFn != nil {
		return m.stopAppFn(ctx, name)
	}
	return 0, nil
}

func (m *mockTruenasClient) RestartApp(ctx context.Context, name string) (int, error) {
	if m.restartAppFn != nil {
		return m.restartAppFn(ctx, name)
	}
	return 0, nil
}

func (m *mockTruenasClient) ListImages(ctx context.Context) ([]truenas.Image, error) {
	if m.listImagesFn != nil {
		return m.listImagesFn(ctx)
	}
	return nil, nil
}

func (m *mockTruenasClient) CreateApp(ctx context.Context, params *truenas.CreateAppParams) (int, error) {
	if m.createAppFn != nil {
		return m.createAppFn(ctx, params)
	}
	return 0, nil
}

func (m *mockTruenasClient) DeleteApp(ctx context.Context, name string) error {
	if m.deleteAppFn != nil {
		return m.deleteAppFn(ctx, name)
	}
	return nil
}

func (m *mockTruenasClient) UpgradeApp(ctx context.Context, name, version string) (int, error) {
	if m.upgradeAppFn != nil {
		return m.upgradeAppFn(ctx, name, version)
	}
	return 0, nil
}

func (m *mockTruenasClient) GetUpgradeSummary(ctx context.Context, name string) (*truenas.AppUpgradeSummary, error) {
	if m.getUpgradeSummaryFn != nil {
		return m.getUpgradeSummaryFn(ctx, name)
	}
	return &truenas.AppUpgradeSummary{}, nil
}

func (m *mockTruenasClient) RollbackApp(ctx context.Context, name, version string) (int, error) {
	if m.rollbackAppFn != nil {
		return m.rollbackAppFn(ctx, name, version)
	}
	return 0, nil
}

func (m *mockTruenasClient) ListInterfaces(ctx context.Context) ([]truenas.Interface, error) {
	if m.listInterfacesFn != nil {
		return m.listInterfacesFn(ctx)
	}
	return nil, nil
}

func (m *mockTruenasClient) ListDirectory(ctx context.Context, path string) ([]truenas.DirEntry, error) {
	if m.listDirectoryFn != nil {
		return m.listDirectoryFn(ctx, path)
	}
	return nil, nil
}
