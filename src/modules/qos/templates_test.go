package qos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModuleScaffoldAndLocalTemplates(t *testing.T) {
	root := filepath.Join(t.TempDir(), "qos")
	c := Config{Interface: "vpn7", Rate: "75Mbit"}
	if err := initialize(root, c); err != nil {
		t.Fatal(err)
	}
	got, err := ReadConfig(filepath.Join(root, "channel.json"))
	if err != nil || got != c {
		t.Fatalf("%+v %v", got, err)
	}
	templatePath := filepath.Join(root, "templates", "qos.service.tmpl")
	custom := ServiceTemplate + "\n# locally edited\n"
	if err := os.WriteFile(templatePath, []byte(custom), 0600); err != nil {
		t.Fatal(err)
	}
	if err := initialize(root, Config{Interface: "wg0", Rate: "100Mbit"}); err != nil {
		t.Fatal(err)
	}
	got, err = ReadConfig(filepath.Join(root, "channel.json"))
	if err != nil || got != c {
		t.Fatal("existing config overwritten")
	}
	unitPath := filepath.Join(root, "generated", "qos.service")
	unit, err := os.ReadFile(unitPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(unit), "# locally edited") || !strings.Contains(string(unit), "wg-quick@vpn7.service") || !strings.Contains(string(unit), filepath.Join(root, "channel.json")) {
		t.Fatal("local template not used")
	}
	if err := os.WriteFile(templatePath, []byte("{{ .UnknownField }}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := generate(root, nil); err == nil {
		t.Fatal("accepted unknown template field")
	}
	after, err := os.ReadFile(unitPath)
	if err != nil || string(after) != string(unit) {
		t.Fatal("failed render replaced unit")
	}
	for _, path := range []string{filepath.Join(root, "channel.json"), templatePath} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("permissions: %s %v", path, err)
		}
	}
}

func TestInvalidConfigTemplateDoesNotCreateChannel(t *testing.T) {
	root := filepath.Join(t.TempDir(), "qos")
	dir := filepath.Join(root, "templates")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "channel.json.tmpl"), []byte(`{"interface":"wg0","rate":"1Gbit"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := initialize(root, Config{Interface: "vpn7", Rate: "75Mbit"}); err == nil {
		t.Fatal("template silently changed requested limit")
	}
	if _, err := os.Stat(filepath.Join(root, "channel.json")); !os.IsNotExist(err) {
		t.Fatal("unexpected generated config")
	}
}
