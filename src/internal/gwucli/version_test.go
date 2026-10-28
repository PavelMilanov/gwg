package gwucli

import (
	"regexp"
	"strings"
	"testing"
)

func TestEmbeddedVersion(t *testing.T) {
	if Version != strings.TrimSpace(versionText) {
		t.Fatalf("version %q does not match VERSION", Version)
	}
	if !regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).MatchString(Version) {
		t.Fatalf("invalid component version %q", Version)
	}
}
