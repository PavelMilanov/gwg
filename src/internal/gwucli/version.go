package gwucli

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionText string

// Version is shared by the CLI and package build through the VERSION file.
var Version = strings.TrimSpace(versionText)
