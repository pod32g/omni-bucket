package version

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		commit   string
		date     string
		modified bool
		want     string
	}{
		{name: "empty is dev", want: "dev"},
		{name: "version only", version: "v0.1.0", want: "v0.1.0"},
		{
			name:    "version with commit and date",
			version: "v0.1.0", commit: "a1b2c3d4e5f6a7b8", date: "2026-06-10T12:00:00Z",
			want: "v0.1.0 (a1b2c3d4e5f6, 2026-06-10T12:00:00Z)",
		},
		{
			name:   "dirty commit, no version falls back to dev",
			commit: "a1b2c3d4e5f6", modified: true,
			want: "dev (a1b2c3d4e5f6-dirty)",
		},
		{
			name:    "commit truncated to 12 chars",
			version: "v1", commit: "0123456789abcdef",
			want: "v1 (0123456789ab)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := format(tt.version, tt.commit, tt.date, tt.modified); got != tt.want {
				t.Fatalf("format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInfoNeverEmpty(t *testing.T) {
	if Info() == "" {
		t.Fatal("Info() should never return an empty string")
	}
}
