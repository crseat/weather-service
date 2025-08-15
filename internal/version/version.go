// Package version holds build-time version metadata.
package version

var (
	// Version is the semantic version of the build (set via -ldflags).
	Version = "dev"
	// Commit is the git commit hash for the build (set via -ldflags).
	Commit = "none"
	// BuiltAt is the build timestamp (set via -ldflags).
	BuiltAt = "unknown"
)
