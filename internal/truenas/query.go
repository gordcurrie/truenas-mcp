package truenas

import (
	"encoding/json"
	"fmt"
	"net/url"
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
// to append to a list endpoint.
//
// filter, if non-empty, is serialised as the TrueNAS query-filters parameter.
// opts.Limit and opts.Offset are included as a query-options JSON object when
// either is non-zero.
//
// Returns ("", nil) when neither filter nor opts fields are set.
func buildQueryString(filter [][]string, opts ListOptions) (string, error) {
	var parts []string

	if len(filter) > 0 {
		f, err := json.Marshal(filter)
		if err != nil {
			return "", fmt.Errorf("marshalling query filter: %w", err)
		}
		parts = append(parts, "query-filters="+url.QueryEscape(string(f)))
	}

	if opts.Limit > 0 || opts.Offset > 0 {
		qo, err := json.Marshal(map[string]int{
			"limit":  opts.Limit,
			"offset": opts.Offset,
		})
		if err != nil {
			return "", fmt.Errorf("marshalling query options: %w", err)
		}
		parts = append(parts, "query-options="+url.QueryEscape(string(qo)))
	}

	if len(parts) == 0 {
		return "", nil
	}
	return "?" + strings.Join(parts, "&"), nil
}
