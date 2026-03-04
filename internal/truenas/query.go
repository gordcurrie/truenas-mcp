package truenas

import (
	"net/url"
	"strconv"
	"strings"
)

// ListOptions controls optional server-side pagination for TrueNAS list endpoints.
// Zero values mean "return all results" (no limit, no offset).
type ListOptions struct {
	// Limit caps the number of results returned by the server. Zero means no limit.
	Limit int
	// Offset skips the first N results. Zero means start from the beginning.
	Offset int
}

// buildQueryString constructs the URL query string (including the leading "?")
// to append to a TrueNAS list endpoint.
//
// TrueNAS SCALE REST API accepts filters and pagination as plain query
// parameters — e.g. ?dataset=Storage&limit=10&offset=0. The JSON-encoded
// query-filters / query-options blob format is not supported by GET endpoints.
//
// filter elements must be three-element slices [field, "=", value]; only
// equality filters are used in this codebase. Non-equality elements are
// silently skipped.
//
// Returns "" when no filter fields are set and opts is zero.
func buildQueryString(filter [][]string, opts ListOptions) string {
	var parts []string

	for _, f := range filter {
		if len(f) == 3 && f[1] == "=" && f[0] != "" {
			parts = append(parts, url.QueryEscape(f[0])+"="+url.QueryEscape(f[2]))
		}
	}

	if opts.Limit > 0 {
		parts = append(parts, "limit="+strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		parts = append(parts, "offset="+strconv.Itoa(opts.Offset))
	}

	if len(parts) == 0 {
		return ""
	}
	return "?" + strings.Join(parts, "&")
}
