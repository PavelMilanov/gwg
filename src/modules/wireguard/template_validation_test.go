package wireguard

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestNativeNATTemplate(t *testing.T) {
	configureTestDirectories(t)
	for _, alias := range []string{"wg0", "vpn7", "wg-1", "wg.1", "wg+1", "wg=1"} {
		t.Run(alias, func(t *testing.T) {
			config := validTestServerConfig()
			config.Alias, config.Eth = alias, "eth+1"
			active := validTestUserConfig("alice", "10.7.0.2/32")
			blocked := active
			blocked.Name, blocked.Status = "bob", ""
			config.Users = []UserConfig{active, blocked}
			data, err := serverTemplate.render(config)
			if err != nil {
				t.Fatal(err)
			}
			s := string(data)
			table, err := natTableName(alias)
			if err != nil {
				t.Fatal(err)
			}
			for _, required := range []string{"nft 'create table ip " + table, "priority srcnat", `oifname "eth+1" counter masquerade`, "PostDown = nft delete table ip " + table, "[Peer]\n# Name = alice\nPublicKey = ", "# [Peer]\n# Name = bob\n# PublicKey = "} {
				if !strings.Contains(s, required) {
					t.Errorf("missing %q", required)
				}
			}
			for _, forbidden := range []string{"iptables", "ip rule", "FORWARD", "flush", "ip saddr"} {
				if strings.Contains(s, forbidden) {
					t.Errorf("unexpected %q", forbidden)
				}
			}
		})
	}
}

func TestNATTableNamesAreSafeAndDistinct(t *testing.T) {
	seen := map[string]bool{}
	for _, alias := range []string{"wg0", "wg-1", "wg.1", "wg+1", "wg=1", "wg_1"} {
		name, err := natTableName(alias)
		if err != nil {
			t.Fatal(err)
		}
		if !regexp.MustCompile(`^gwu_nat_[0-9a-f]+$`).MatchString(name) || seen[name] {
			t.Fatalf("invalid or duplicate name %q", name)
		}
		seen[name] = true
	}
	if _, err := natTableName("wg0; touch /tmp/x"); err == nil {
		t.Fatal("accepted unsafe alias")
	}
}

func TestExistingServerTemplateRequiresExplicitUpdate(t *testing.T) {
	configureTestDirectories(t)
	old := []byte("PostUp = iptables -A FORWARD -i %i -j ACCEPT\n")
	if err := WriteConfigTemplate("server", old); err != nil {
		t.Fatal(err)
	}
	if err := ensureConfigTemplates(); err != nil {
		t.Fatal(err)
	}
	data, err := ReadConfigTemplate("server")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(old) {
		t.Fatal("existing template was automatically migrated")
	}
	if err := ResetConfigTemplate("server"); err != nil {
		t.Fatal(err)
	}
	data, err = ReadConfigTemplate("server")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != defaultServerTemplate {
		t.Fatal("explicit reset did not install new template")
	}
	if _, err := os.Stat(configPath(serverDir, "wg0", ".conf")); !os.IsNotExist(err) {
		t.Fatal("template update unexpectedly generated WG config")
	}
}

func TestTemplateValidationPreservesPreviousFile(t *testing.T) {
	configureTestDirectories(t)
	previous := []byte("# previous template\n")
	if err := WriteConfigTemplate("server", previous); err != nil {
		t.Fatal(err)
	}
	invalid := []string{
		"{{ .UnknownField }}",
		"{{ range .Users }}{{ if .Status }}{{ .UnknownActive }}{{ end }}{{ end }}",
		"{{ range .Users }}{{ if not .Status }}{{ .UnknownBlocked }}{{ end }}{{ end }}",
		"{{ if not .Users }}{{ .UnknownWithoutUsers }}{{ end }}",
		"PostUp = echo '\n",
		"PostDown = if true; then\n",
		"PreUp = echo \"\n",
		"PreDown = (\n",
	}
	for _, source := range invalid {
		if err := WriteConfigTemplate("server", []byte(source)); err == nil {
			t.Errorf("accepted %q", source)
		}
		data, err := os.ReadFile(serverTemplate.path())
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != string(previous) {
			t.Fatal("invalid template replaced previous file")
		}
	}
	if err := WriteConfigTemplate("client", []byte("{{ .UnknownField }}")); err == nil {
		t.Fatal("accepted unknown client field")
	}
}

func TestHookValidationDoesNotExecuteShell(t *testing.T) {
	configureTestDirectories(t)
	marker := filepath.Join(t.TempDir(), "must-not-exist")
	startup := filepath.Join(t.TempDir(), "bash-env")
	if err := os.WriteFile(startup, []byte("touch "+marker+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BASH_ENV", startup)
	for _, hook := range []string{"PostUp = touch " + marker, "PostDown = echo $(touch " + marker + ")", "PreUp = echo `touch " + marker + "`"} {
		if err := WriteConfigTemplate("server", []byte(hook+"\n")); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("shell command executed: %v", err)
	}
}

func TestRenderedHookFailurePreservesConfiguration(t *testing.T) {
	configureTestDirectories(t)
	if err := ensureConfigTemplates(); err != nil {
		t.Fatal(err)
	}
	// Direct vim edits bypass template set; actual rendering must still validate.
	if err := os.WriteFile(serverTemplate.path(), []byte("PostUp = echo '\n"), 0600); err != nil {
		t.Fatal(err)
	}
	path := configPath(serverDir, "wg7", ".conf")
	if err := os.WriteFile(path, []byte("previous\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeServerConfig(validTestServerConfig(), "wg7"); err == nil {
		t.Fatal("accepted malformed hook")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "previous\n" {
		t.Fatal("configuration replaced on validation error")
	}
}

func TestTemplateErrorsDoNotExposePrivateKey(t *testing.T) {
	for _, source := range []string{`{{ call .ServerPrivateKey }}`, "PostUp = echo '{{ .ServerPrivateKey }}\n"} {
		tmpl, err := parseConfigTemplate("server", []byte(source))
		if err != nil {
			t.Fatal(err)
		}
		_, err = renderConfigTemplate(tmpl, validTestServerConfig())
		if err == nil || strings.Contains(err.Error(), testKey) {
			t.Fatalf("unsafe error: %v", err)
		}
	}
}

func TestUserLifecycleKeepsNativeHooks(t *testing.T) {
	configureTestDirectories(t)
	config := validTestServerConfig()
	if err := config.createServerConfigFile(); err != nil {
		t.Fatal(err)
	}
	fake := &fakeRunner{outputs: map[string][]byte{"wg genkey": []byte(testKey + "\n"), "wg pubkey": []byte(testKey + "\n")}}
	previous := commandRunner
	commandRunner = fake
	t.Cleanup(func() { commandRunner = previous })
	steps := []struct {
		name string
		run  func() error
		peer string
	}{
		{"add", func() error { return AddUser("alice") }, "[Peer]\n# Name = alice\nPublicKey = "},
		{"block", func() error { return ChangeStatusUser("alice", "block") }, "# [Peer]\n# Name = alice\n# PublicKey = "},
		{"unblock", func() error { return ChangeStatusUser("alice", "unblock") }, "[Peer]\n# Name = alice\nPublicKey = "},
		{"remove", func() error { return RemoveUser("alice") }, ""},
	}
	for _, step := range steps {
		if err := step.run(); err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
		data, err := os.ReadFile(configPath(serverDir, "wg7", ".conf"))
		if err != nil {
			t.Fatal(err)
		}
		s := string(data)
		if strings.Contains(s, "iptables") || strings.Contains(s, "ip rule") || strings.Count(s, "PostUp = ") != 1 || !strings.Contains(s, "nft 'create table ip gwu_nat_776737") {
			t.Fatalf("%s: hooks changed", step.name)
		}
		if step.peer != "" && !strings.Contains(s, step.peer) {
			t.Fatalf("%s: peer state not updated", step.name)
		}
		if step.peer == "" && strings.Contains(s, "# Name = alice") {
			t.Fatal("removed peer remains")
		}
	}
}
