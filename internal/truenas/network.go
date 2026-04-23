package truenas

import (
	"context"
	"fmt"
)

// InterfaceAlias represents an IP address assignment on a network interface.
type InterfaceAlias struct {
	// Type is the address family: "INET" (IPv4) or "INET6" (IPv6).
	Type    string `json:"type"`
	Address string `json:"address"`
	Netmask int    `json:"netmask"`
}

// Interface represents a network interface on TrueNAS SCALE.
type Interface struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Type is the interface kind: PHYSICAL, BRIDGE, BOND, VLAN, etc.
	Type        string           `json:"type"`
	Description string           `json:"description,omitempty"`
	Aliases     []InterfaceAlias `json:"aliases,omitempty"`
}

// ListInterfaces returns all network interfaces configured on the TrueNAS host.
func (c *Client) ListInterfaces(ctx context.Context) ([]Interface, error) {
	var ifaces []Interface
	if err := c.call(ctx, "interface.query", buildQueryParams(nil, ListOptions{}), &ifaces); err != nil {
		return nil, fmt.Errorf("listing interfaces: %w", err)
	}
	return ifaces, nil
}
