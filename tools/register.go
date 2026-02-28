// Package tools registers all MCP tools onto the server.
package tools

import (
	"github.com/gordcurrie/truenas-mcp/internal/truenas"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config holds optional feature flags for tool registration.
type Config struct {
	// AllowDestructive enables destructive tools (delete_dataset, delete_vm, etc.).
	AllowDestructive bool
}

// RegisterAll wires all TrueNAS MCP tools onto the provided server.
func RegisterAll(s *mcp.Server, client *truenas.Client, cfg Config) {
	registerSystemTools(s, client)
	registerPoolTools(s, client)
	registerDatasetTools(s, client)
	_ = cfg
}
