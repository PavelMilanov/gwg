package layout

import (
	"path/filepath"
	"testing"
)

func TestModulesUseCommonParent(t *testing.T) {
	if BaseDir != "/etc/wireguard/gwu" {
		t.Fatal(BaseDir)
	}
	for _, path := range []string{WireGuardDir, QoSDir} {
		if filepath.Dir(path) != BaseDir {
			t.Fatal(path)
		}
	}
	if WireGuardDir == QoSDir {
		t.Fatal("module paths overlap")
	}
}
