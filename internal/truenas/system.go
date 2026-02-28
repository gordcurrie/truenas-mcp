package truenas

import (
	"context"
	"fmt"
)

// GetSystemInfo returns general system information from GET /system/info.
func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	var info SystemInfo
	if err := c.get(ctx, "/system/info", &info); err != nil {
		return nil, fmt.Errorf("getting system info: %w", err)
	}
	return &info, nil
}
