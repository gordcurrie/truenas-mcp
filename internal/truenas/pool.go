package truenas

import (
	"context"
	"fmt"
)

// PoolScan holds the result of the most recent scrub or resilver on a pool.
type PoolScan struct {
	Function   string  `json:"function"`
	State      string  `json:"state"`
	Errors     int     `json:"errors"`
	Percentage float64 `json:"percentage"`
}

// Pool represents a ZFS storage pool on TrueNAS SCALE.
type Pool struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	GUID          string   `json:"guid"`
	Status        string   `json:"status"`
	Path          string   `json:"path"`
	Healthy       bool     `json:"healthy"`
	Warning       bool     `json:"warning"`
	StatusCode    string   `json:"status_code"`
	StatusDetail  *string  `json:"status_detail"`
	Size          int64    `json:"size"`
	Allocated     int64    `json:"allocated"`
	Free          int64    `json:"free"`
	Freeing       int64    `json:"freeing"`
	Fragmentation string   `json:"fragmentation"`
	Scan          PoolScan `json:"scan"`
}

// ListPools returns all ZFS pools on the system.
// Pass a ListOptions value to apply server-side pagination (limit / offset).
func (c *Client) ListPools(ctx context.Context, opts ...ListOptions) ([]Pool, error) {
	var o ListOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	qs, err := buildQueryString(nil, o)
	if err != nil {
		return nil, fmt.Errorf("listing pools: %w", err)
	}
	var pools []Pool
	if err := c.get(ctx, "/pool"+qs, &pools); err != nil {
		return nil, fmt.Errorf("listing pools: %w", err)
	}
	return pools, nil
}

// GetPool returns a single ZFS pool by its numeric ID.
func (c *Client) GetPool(ctx context.Context, id int) (*Pool, error) {
	var pool Pool
	if err := c.get(ctx, fmt.Sprintf("/pool/id/%d", id), &pool); err != nil {
		return nil, fmt.Errorf("getting pool %d: %w", id, err)
	}
	return &pool, nil
}
