package truenas

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a requested TrueNAS resource does not exist.
var ErrNotFound = errors.New("resource not found")

// ErrMethodNotFound is returned when the TrueNAS API returns JSON-RPC -32601
// (method not found). This typically means an older TrueNAS build that does not
// implement the requested RPC method.
var ErrMethodNotFound = errors.New("RPC method not found")

// APIError represents a non-2xx response from the TrueNAS API.
type APIError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("TrueNAS API error %d: %s", e.StatusCode, e.Body)
}

// SystemInfo represents the response from GET /system/info.
type SystemInfo struct {
	Version       string    `json:"version"`
	Hostname      string    `json:"hostname"`
	PhysMem       int64     `json:"physmem"`
	Model         string    `json:"model"`
	Cores         int       `json:"cores"`
	LoadAvg       []float64 `json:"loadavg"`
	UptimeSeconds float64   `json:"uptime_seconds"`
	Timezone      string    `json:"timezone"`
	SystemProduct string    `json:"system_product"`
	SystemSerial  string    `json:"system_serial"`
	Manufacturer  string    `json:"manufacturer"`
}
