// Package version reports the bb build version.
package version

import "runtime/debug"

// Version is injected at build time via
// -ldflags "-X github.com/pod32g/omni-bucket/internal/version.Version=<v>".
// When empty, Info falls back to the VCS data the Go toolchain embeds.
var Version = ""

// Info returns a human-readable version string combining the injected version
// (or the module version when installed via `go install`) with the VCS commit,
// dirty flag, and build date.
func Info() string {
	v := Version
	var commit, date string
	var modified bool
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v == "" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			v = bi.Main.Version
		}
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				commit = s.Value
			case "vcs.time":
				date = s.Value
			case "vcs.modified":
				modified = s.Value == "true"
			}
		}
	}
	return format(v, commit, date, modified)
}

// format renders the version string. A missing version is reported as "dev";
// the commit is truncated to 12 chars and annotated with -dirty when modified.
func format(version, commit, date string, modified bool) string {
	if version == "" {
		version = "dev"
	}
	if commit == "" {
		return version
	}
	if len(commit) > 12 {
		commit = commit[:12]
	}
	out := version + " (" + commit
	if modified {
		out += "-dirty"
	}
	if date != "" {
		out += ", " + date
	}
	return out + ")"
}
