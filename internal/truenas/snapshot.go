package truenas

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// Snapshot represents a ZFS snapshot on TrueNAS SCALE.
// The ID field is the full snapshot identifier in "dataset@name" form.
type Snapshot struct {
	// ID is the full snapshot identifier, e.g. "Storage/backups@before-upgrade".
	ID         string   `json:"id"`
	Dataset    string   `json:"dataset"`
	Name       string   `json:"snapshot_name"`
	Pool       string   `json:"pool"`
	Referenced string   `json:"referenced"`
	Used       string   `json:"used"`
	CreateTXG  string   `json:"createtxg"`
	GUID       string   `json:"guid"`
	Clones     []string `json:"clones"`
}

// CreateSnapshotParams holds the fields for creating a ZFS snapshot via
// POST /pool/snapshot.
type CreateSnapshotParams struct {
	// Dataset is the full dataset path, e.g. "Storage/backups". Required.
	Dataset string `json:"dataset"`
	// Name is the snapshot suffix that appears after the @ sign. Required.
	Name string `json:"name"`
	// Recursive, when true, also creates snapshots on all descendant datasets.
	Recursive bool `json:"recursive,omitempty"`
}

// RollbackSnapshotParams holds options for rolling a dataset back to a snapshot.
type RollbackSnapshotParams struct {
	// Recursive destroys any more recent snapshot on the dataset and its
	// descendants before rolling back.
	Recursive bool `json:"recursive,omitempty"`
	// RecursiveClones also destroys more recent clones when set alongside Recursive.
	RecursiveClones bool `json:"recursive_clones,omitempty"`
	// Force forces rollback even if the dataset is busy.
	Force bool `json:"force,omitempty"`
}

// snapshotIDPath returns the URL path for a snapshot ID resource.
// id is the full "dataset@snapname" identifier.
func snapshotIDPath(id string) string {
	return "/pool/snapshot/id/" + url.PathEscape(id)
}

// ListSnapshots returns ZFS snapshots visible to the API.
// If dataset is non-empty, a server-side query filter is applied so that only
// snapshots for that dataset are transferred over the wire. This avoids fetching
// potentially thousands of snapshots when only one dataset's history is needed.
// Pass a ListOptions value to apply server-side pagination (limit / offset).
func (c *Client) ListSnapshots(ctx context.Context, dataset string, opts ...ListOptions) ([]Snapshot, error) {
	var o ListOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	if err := validateListOptions(o); err != nil {
		return nil, fmt.Errorf("listing snapshots: %w", err)
	}
	var filter [][]string
	if dataset != "" {
		filter = [][]string{{"dataset", "=", dataset}}
	}
	qs := buildQueryString(filter, o)
	var snapshots []Snapshot
	if err := c.get(ctx, "/pool/snapshot"+qs, &snapshots); err != nil {
		return nil, fmt.Errorf("listing snapshots: %w", err)
	}
	return snapshots, nil
}

// GetSnapshot returns a single snapshot by its full ID (e.g. "Storage/backups@before-upgrade").
func (c *Client) GetSnapshot(ctx context.Context, id string) (*Snapshot, error) {
	if dataset, name, ok := strings.Cut(id, "@"); !ok || dataset == "" || name == "" || strings.Contains(name, "@") {
		return nil, fmt.Errorf("getting snapshot: id must be in dataset@name form with non-empty dataset and name, got %q", id)
	}

	var snap Snapshot
	if err := c.get(ctx, snapshotIDPath(id), &snap); err != nil {
		return nil, fmt.Errorf("getting snapshot %q: %w", id, err)
	}
	return &snap, nil
}

// CreateSnapshot creates a new ZFS snapshot and returns the created Snapshot.
// Both params.Dataset and params.Name are required.
func (c *Client) CreateSnapshot(ctx context.Context, params CreateSnapshotParams) (*Snapshot, error) {
	if params.Dataset == "" {
		return nil, fmt.Errorf("creating snapshot: dataset must not be empty")
	}
	if params.Name == "" {
		return nil, fmt.Errorf("creating snapshot: name must not be empty")
	}

	var snap Snapshot
	if err := c.postWithBody(ctx, "/pool/snapshot", params, &snap); err != nil {
		return nil, fmt.Errorf("creating snapshot %q@%q: %w", params.Dataset, params.Name, err)
	}
	return &snap, nil
}

// RollbackSnapshot rolls the dataset back to the specified snapshot.
// id must be the full "dataset@name" identifier.
// The API call is synchronous; an error is returned only if the request fails.
func (c *Client) RollbackSnapshot(ctx context.Context, id string, params RollbackSnapshotParams) error {
	if dataset, name, ok := strings.Cut(id, "@"); !ok || dataset == "" || name == "" || strings.Contains(name, "@") {
		return fmt.Errorf("rolling back snapshot: id must be in dataset@name form with non-empty dataset and name, got %q", id)
	}

	path := snapshotIDPath(id) + "/rollback"
	if err := c.postWithBody(ctx, path, params, nil); err != nil {
		return fmt.Errorf("rolling back snapshot %q: %w", id, err)
	}
	return nil
}

// DeleteSnapshot deletes the ZFS snapshot identified by id (e.g. "Storage/backups@before-upgrade").
func (c *Client) DeleteSnapshot(ctx context.Context, id string) error {
	if dataset, name, ok := strings.Cut(id, "@"); !ok || dataset == "" || name == "" || strings.Contains(name, "@") {
		return fmt.Errorf("deleting snapshot: id must be in dataset@name form with non-empty dataset and name, got %q", id)
	}

	if err := c.delete(ctx, snapshotIDPath(id)); err != nil {
		return fmt.Errorf("deleting snapshot %q: %w", id, err)
	}
	return nil
}
