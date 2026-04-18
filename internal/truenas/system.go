package truenas

import (
	"context"
	"fmt"
)

// GetSystemInfo returns general system information via the system.info RPC method.
func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	var info SystemInfo
	if err := c.call(ctx, "system.info", []any{}, &info); err != nil {
		return nil, fmt.Errorf("getting system info: %w", err)
	}
	return &info, nil
}
