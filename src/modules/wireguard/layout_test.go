package wireguard

import (
	"os"
	"strings"
	"testing"

	"github.com/PavelMilanov/go-wg-manager/internal/layout"
)

func TestModulePaths(t *testing.T) {
	for _, path := range []string{managerDir, userConfigDir, usersDir, serverTemplate.path(), clientTemplate.path(), gatewayTemplate.path()} {
		if !strings.HasPrefix(path, layout.WireGuardDir+"/") && path != layout.WireGuardDir {
			t.Fatalf("path outside module: %s", path)
		}
	}
}

func TestInstallRefusesExistingInterfaceConfig(t *testing.T) {
	configureTestDirectories(t)
	path := configPath(serverDir, "wg7", ".conf")
	if err := os.WriteFile(path, []byte("existing production\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := InstallServer("wg7", "10.7.0.1/24", 51830); err == nil {
		t.Fatal("existing config overwritten")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "existing production\n" {
		t.Fatal("config changed")
	}
}

func TestGeneratedInterfaceIsKeptInsideModule(t *testing.T) {
	configureTestDirectories(t)
	c := validTestServerConfig()
	if err := writeServerConfig(c, c.Alias); err != nil {
		t.Fatal(err)
	}
	generated, err := os.ReadFile(configPath(managerDir+"/generated", c.Alias, ".conf"))
	if err != nil {
		t.Fatal(err)
	}
	live, err := os.ReadFile(configPath(serverDir, c.Alias, ".conf"))
	if err != nil {
		t.Fatal(err)
	}
	if string(generated) != string(live) {
		t.Fatal("generated/live files diverged")
	}
}
