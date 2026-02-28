package truenas

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a requested resource does not exist (HTTP 404).
var ErrNotFound = errors.New("resource not found")

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
