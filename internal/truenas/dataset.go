package truenas

import (
	"context"
	"fmt"
	"net/url"
)

// DatasetValue is a TrueNAS property wrapper returned for most dataset fields
// (size, quota, compression, etc.). The human-readable string is in Value;
// the raw numeric or string representation is in RawValue; and the parsed Go
// value (int64, string, or nil) is in Parsed.
type DatasetValue struct {
	Value    string `json:"value"`
	RawValue string `json:"rawvalue"`
	Parsed   any    `json:"parsed"`
}

// Dataset represents a ZFS dataset or zvol on TrueNAS SCALE.
type Dataset struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Pool        string       `json:"pool"`
	Type        string       `json:"type"` // FILESYSTEM or VOLUME
	MountPoint  *string      `json:"mountpoint"`
	Encrypted   bool         `json:"encrypted"`
	Available   DatasetValue `json:"available"`
	Used        DatasetValue `json:"used"`
	Quota       DatasetValue `json:"quota"`
	Compression DatasetValue `json:"compression"`
	Comments    DatasetValue `json:"comments"`
	Children    []Dataset    `json:"children"`
}

// CreateDatasetParams holds the fields required and optional for creating a
// new ZFS dataset via POST /pool/dataset.
type CreateDatasetParams struct {
	// Name is the full dataset path including the pool, e.g. "Storage/backups".
	// This field is required.
	Name string `json:"name"`
	// Type is "FILESYSTEM" (default) or "VOLUME".
	Type string `json:"type,omitempty"`
	// Compression algorithm, e.g. "lz4", "gzip", "zstd". Omit to use pool default.
	Compression string `json:"compression,omitempty"`
	// Comments is a free-text description for the dataset.
	Comments string `json:"comments,omitempty"`
	// Quota is the maximum space (in bytes) allowed for this dataset and its
	// children. Zero or omitted means no quota.
	Quota int64 `json:"quota,omitempty"`
}

// ListDatasets returns all ZFS datasets and zvols visible to the API.
// If pool is non-empty, only datasets whose Pool field matches are returned.
func (c *Client) ListDatasets(ctx context.Context, pool string) ([]Dataset, error) {
	var datasets []Dataset
	if err := c.get(ctx, "/pool/dataset", &datasets); err != nil {
		return nil, fmt.Errorf("listing datasets: %w", err)
	}

	if pool == "" {
		return datasets, nil
	}

	filtered := make([]Dataset, 0, len(datasets))
	for i := range datasets {
		if datasets[i].Pool == pool {
			filtered = append(filtered, datasets[i])
		}
	}
	return filtered, nil
}

// GetDataset returns a single dataset by its full path ID (e.g. "Storage/backups").
// The ID is URL-path-encoded before being inserted into the request URL so that
// slashes in the path are transmitted correctly.
func (c *Client) GetDataset(ctx context.Context, id string) (*Dataset, error) {
	var dataset Dataset
	path := "/pool/dataset/id/" + url.PathEscape(id)
	if err := c.get(ctx, path, &dataset); err != nil {
		return nil, fmt.Errorf("getting dataset %q: %w", id, err)
	}
	return &dataset, nil
}

// CreateDataset creates a new ZFS dataset with the provided parameters and
// returns the newly created Dataset. The caller must supply at least params.Name.
func (c *Client) CreateDataset(ctx context.Context, params CreateDatasetParams) (*Dataset, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("creating dataset: name must not be empty")
	}

	var dataset Dataset
	if err := c.postWithBody(ctx, "/pool/dataset", params, &dataset); err != nil {
		return nil, fmt.Errorf("creating dataset %q: %w", params.Name, err)
	}
	return &dataset, nil
}
