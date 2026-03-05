package truenas

import (
	"testing"
)

func TestBuildQueryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		filter [][]string
		opts   ListOptions
		want   string
	}{
		{
			name: "empty",
			want: "",
		},
		{
			name:   "filter only – simple value",
			filter: [][]string{{"pool", "=", "Storage"}},
			want:   "?pool=Storage",
		},
		{
			name:   "filter only – slash in value is percent-encoded",
			filter: [][]string{{"dataset", "=", "Storage/backups"}},
			want:   "?dataset=Storage%2Fbackups",
		},
		{
			name: "limit only",
			opts: ListOptions{Limit: 5},
			want: "?limit=5",
		},
		{
			name: "offset only",
			opts: ListOptions{Offset: 2},
			want: "?offset=2",
		},
		{
			name: "limit and offset",
			opts: ListOptions{Limit: 10, Offset: 3},
			want: "?limit=10&offset=3",
		},
		{
			name:   "filter with limit and offset",
			filter: [][]string{{"pool", "=", "Storage"}},
			opts:   ListOptions{Limit: 5, Offset: 1},
			want:   "?pool=Storage&limit=5&offset=1",
		},
		{
			name: "zero limit is omitted",
			opts: ListOptions{Limit: 0, Offset: 2},
			want: "?offset=2",
		},
		{
			name: "zero offset is omitted",
			opts: ListOptions{Limit: 3, Offset: 0},
			want: "?limit=3",
		},
		{
			name: "both zero – no params emitted",
			opts: ListOptions{Limit: 0, Offset: 0},
			want: "",
		},
		{
			name: "non-equality filter is skipped",
			filter: [][]string{
				{"pool", "!=", "Storage"},
				{"name", "=", "backups"},
			},
			want: "?name=backups",
		},
		{
			name:   "multiple equality filters",
			filter: [][]string{{"pool", "=", "apps"}, {"type", "=", "FILESYSTEM"}},
			want:   "?pool=apps&type=FILESYSTEM",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := buildQueryString(tc.filter, tc.opts)
			if got != tc.want {
				t.Errorf("buildQueryString(%v, %+v) = %q, want %q", tc.filter, tc.opts, got, tc.want)
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
		{"zero values", ListOptions{}, false},
		{"positive limit only", ListOptions{Limit: 10}, false},
		{"positive offset only", ListOptions{Offset: 5}, false},
		{"positive limit and offset", ListOptions{Limit: 10, Offset: 5}, false},
		{"negative limit", ListOptions{Limit: -1}, true},
		{"negative offset", ListOptions{Offset: -1}, true},
		{"large negative limit", ListOptions{Limit: -100}, true},
		{"negative limit zero offset", ListOptions{Limit: -1, Offset: 0}, true},
		{"zero limit negative offset", ListOptions{Limit: 0, Offset: -3}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateListOptions(tc.opts)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateListOptions(%+v) error = %v, wantErr = %v", tc.opts, err, tc.wantErr)
			}
		})
	}
}
