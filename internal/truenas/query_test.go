package truenas

import (
	"encoding/json"
	"testing"
)

func TestBuildQueryParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		filters [][]string
		opts    ListOptions
		// wantFilters is the JSON encoding of the first element
		wantFilters string
		// wantOptions is the JSON encoding of the second element
		wantOptions string
	}{
		{
			name:        "empty – no filters, no opts",
			wantFilters: "[]",
			wantOptions: "{}",
		},
		{
			name:        "single filter",
			filters:     [][]string{{"pool", "=", "Storage"}},
			wantFilters: `[["pool","=","Storage"]]`,
			wantOptions: "{}",
		},
		{
			name:        "filter with slash in value",
			filters:     [][]string{{"dataset", "=", "Storage/backups"}},
			wantFilters: `[["dataset","=","Storage/backups"]]`,
			wantOptions: "{}",
		},
		{
			name:        "limit only",
			opts:        ListOptions{Limit: 5},
			wantFilters: "[]",
			wantOptions: `{"limit":5}`,
		},
		{
			name:        "offset only",
			opts:        ListOptions{Offset: 2},
			wantFilters: "[]",
			wantOptions: `{"offset":2}`,
		},
		{
			name:        "limit and offset",
			opts:        ListOptions{Limit: 10, Offset: 3},
			wantFilters: "[]",
			wantOptions: `{"limit":10,"offset":3}`,
		},
		{
			name:        "filter with limit and offset",
			filters:     [][]string{{"pool", "=", "Storage"}},
			opts:        ListOptions{Limit: 5, Offset: 1},
			wantFilters: `[["pool","=","Storage"]]`,
			wantOptions: `{"limit":5,"offset":1}`,
		},
		{
			name:        "zero limit is omitted",
			opts:        ListOptions{Limit: 0, Offset: 2},
			wantFilters: "[]",
			wantOptions: `{"offset":2}`,
		},
		{
			name:        "zero offset is omitted",
			opts:        ListOptions{Limit: 3, Offset: 0},
			wantFilters: "[]",
			wantOptions: `{"limit":3}`,
		},
		{
			name:        "multiple filters",
			filters:     [][]string{{"pool", "=", "apps"}, {"type", "=", "FILESYSTEM"}},
			wantFilters: `[["pool","=","apps"],["type","=","FILESYSTEM"]]`,
			wantOptions: "{}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := buildQueryParams(tc.filters, tc.opts)
			if len(got) != 2 {
				t.Fatalf("buildQueryParams returned %d elements, want 2", len(got))
			}

			filtersJSON, err := json.Marshal(got[0])
			if err != nil {
				t.Fatalf("marshalling filters: %v", err)
			}
			optionsJSON, err := json.Marshal(got[1])
			if err != nil {
				t.Fatalf("marshalling options: %v", err)
			}

			if string(filtersJSON) != tc.wantFilters {
				t.Errorf("filters = %s, want %s", filtersJSON, tc.wantFilters)
			}
			if string(optionsJSON) != tc.wantOptions {
				t.Errorf("options = %s, want %s", optionsJSON, tc.wantOptions)
			}
		})
	}
}

func TestValidateListOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    ListOptions
		wantErr bool
	}{
		{name: "zero values ok", opts: ListOptions{}, wantErr: false},
		{name: "positive limit ok", opts: ListOptions{Limit: 10}, wantErr: false},
		{name: "positive offset ok", opts: ListOptions{Offset: 5}, wantErr: false},
		{name: "negative limit errors", opts: ListOptions{Limit: -1}, wantErr: true},
		{name: "negative offset errors", opts: ListOptions{Offset: -1}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateListOptions(tc.opts)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateListOptions() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
