package truenas

import "fmt"

// ListOptions controls optional server-side pagination for TrueNAS list endpoints.
// Zero values mean "return all results" (no limit, no offset).
type ListOptions struct {
	// Limit caps the number of results returned by the server. Zero means no limit.
	Limit int
	// Offset skips the first N results. Zero means start from the beginning.
	Offset int
}

// validateListOptions returns an error if opts contains negative pagination values.
// Callers must validate before passing opts to buildQueryParams so that negative
// values are caught early rather than silently treated as "no constraint".
func validateListOptions(opts ListOptions) error {
	if opts.Limit < 0 {
		return fmt.Errorf("limit must not be negative, got %d", opts.Limit)
	}
	if opts.Offset < 0 {
		return fmt.Errorf("offset must not be negative, got %d", opts.Offset)
	}
	return nil
}

// buildQueryParams constructs the JSON-RPC params array for TrueNAS query methods.
// It returns [filters, options] where filters is a []any of [field, op, value] triples
// and options is a map containing limit/offset when non-zero.
//
// Example:
//
//	buildQueryParams([][]string{{"pool","=","tank"}}, ListOptions{Limit:10})
//	→ [[["pool","=","tank"]], {"limit":10}]
func buildQueryParams(filters [][]string, opts ListOptions) []any {
	f := make([]any, 0, len(filters))
	for _, filter := range filters {
		if len(filter) == 3 {
			f = append(f, []any{filter[0], filter[1], filter[2]})
		}
	}

	options := map[string]any{}
	if opts.Limit > 0 {
		options["limit"] = opts.Limit
	}
	if opts.Offset > 0 {
		options["offset"] = opts.Offset
	}

	return []any{f, options}
}
