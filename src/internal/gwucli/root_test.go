package gwucli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCommandHierarchy(t *testing.T) {
	rootCmd := NewCommand()
	paths := [][]string{
		{"wireguard", "init"},
		{"wireguard", "server", "install"},
		{"wireguard", "user", "add"},
		{"wireguard", "user", "remove"},
		{"wireguard", "user", "block"},
		{"wireguard", "user", "unblock"},
		{"wireguard", "user", "list"},
		{"wireguard", "template", "show"},
		{"wireguard", "template", "set"},
		{"wireguard", "template", "reset"},
		{"qos", "apply"},
		{"qos", "init"},
		{"qos", "render"},
		{"qos", "show"},
		{"qos", "down"},
		{"version"},
	}
	for _, path := range paths {
		command, remaining, err := rootCmd.Find(path)
		if err != nil {
			t.Errorf("find %q: %v", strings.Join(path, " "), err)
			continue
		}
		if len(remaining) != 0 || command.Name() != path[len(path)-1] {
			t.Errorf("find %q returned %q and remaining %v", strings.Join(path, " "), command.Name(), remaining)
		}
	}
}

func TestUnknownCommandIsRejected(t *testing.T) {
	rootCmd := NewCommand()
	rootCmd.SetArgs([]string{"add"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	if _, err := rootCmd.ExecuteC(); err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("error = %v", err)
	}
}

func TestRemovedServerCommandsAreRejected(t *testing.T) {
	for _, name := range []string{"show", "stat"} {
		t.Run(name, func(t *testing.T) {
			root := NewCommand()
			root.SetArgs([]string{"wireguard", "server", name})
			root.SetOut(&bytes.Buffer{})
			root.SetErr(&bytes.Buffer{})
			if _, err := root.ExecuteC(); err == nil {
				t.Fatalf("removed command server %s was accepted", name)
			}
			server, _, err := NewCommand().Find([]string{"server"})
			if err != nil {
				t.Fatal(err)
			}
			for _, command := range server.Commands() {
				if command.Name() == name {
					t.Fatalf("removed command %s remains registered", name)
				}
			}
		})
	}
}

func TestRoutesCLIIsNotPartOfGWU(t *testing.T) {
	rootCmd := NewCommand()
	rootCmd.SetArgs([]string{"routes"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	if _, err := rootCmd.ExecuteC(); err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("separate routes CLI unexpectedly accepted: %v", err)
	}
}

func TestUnsafeUserNameIsRejectedBeforeFilesystemAccess(t *testing.T) {
	rootCmd := NewCommand()
	rootCmd.SetArgs([]string{"wireguard", "user", "add", "../../root"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	if _, err := rootCmd.ExecuteC(); err == nil || !strings.Contains(err.Error(), "invalid user name") {
		t.Fatalf("error = %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	rootCmd := NewCommand()
	var output bytes.Buffer
	rootCmd.SetArgs([]string{"version"})
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&bytes.Buffer{})
	if _, err := rootCmd.ExecuteC(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "gwu version: "+Version) || !strings.Contains(output.String(), "go version:") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestCommandTreesHaveIndependentFlags(t *testing.T) {
	tests := []struct {
		path  []string
		flag  string
		value string
	}{
		{[]string{"wireguard", "server", "install"}, "name", "vpn7"},
		{[]string{"wireguard", "server", "install"}, "network", "10.7.0.1/24"},
		{[]string{"wireguard", "server", "install"}, "port", "51900"},
		{[]string{"wireguard", "template", "set"}, "file", "custom.tmpl"},
		{[]string{"qos", "apply"}, "config", "/tmp/channel.json"},
		{[]string{"qos", "init"}, "rate", "75Mbit"},
		{[]string{"qos", "show"}, "interface", "vpn7"},
		{[]string{"qos", "down"}, "interface", "vpn7"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.path, "/")+"/"+tt.flag, func(t *testing.T) {
			first, _, err := NewCommand().Find(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			second, _, err := NewCommand().Find(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if err := first.Flags().Set(tt.flag, tt.value); err != nil {
				t.Fatal(err)
			}
			flag := second.Flags().Lookup(tt.flag)
			if flag.Changed || flag.Value.String() != flag.DefValue {
				t.Fatalf("new command inherited flag %s: %s", tt.flag, flag.Value)
			}
		})
	}
}

func TestRemovedTCCommandsAreRejected(t *testing.T) {
	for _, path := range [][]string{{"tc", "service", "up"}, {"tc", "bandwidth", "add", "x"}, {"tc", "filter", "add", "x"}} {
		cmd := NewCommand()
		cmd.SetArgs(path)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		if _, err := cmd.ExecuteC(); err == nil {
			t.Fatalf("accepted obsolete command %v", path)
		}
	}
}

func TestQoSInitRequiresExplicitRate(t *testing.T) {
	cmd := NewCommand()
	cmd.SetArgs([]string{"qos", "init"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if _, err := cmd.ExecuteC(); err == nil || !strings.Contains(err.Error(), "required flag") {
		t.Fatalf("error = %v", err)
	}
}

func TestRequiredFlagsDoNotLeakAcrossTrees(t *testing.T) {
	for _, tt := range []struct {
		path        []string
		flag, value string
		args        []string
	}{
		{[]string{"qos", "init"}, "rate", "75Mbit", []string{"qos", "init"}},
		{[]string{"wireguard", "template", "set"}, "file", "/tmp/custom.tmpl", []string{"wireguard", "template", "set", "server"}},
	} {
		first, _, err := NewCommand().Find(tt.path)
		if err != nil {
			t.Fatal(err)
		}
		if err := first.Flags().Set(tt.flag, tt.value); err != nil {
			t.Fatal(err)
		}
		second := NewCommand()
		second.SetArgs(tt.args)
		second.SetOut(&bytes.Buffer{})
		second.SetErr(&bytes.Buffer{})
		if _, err := second.ExecuteC(); err == nil || !strings.Contains(err.Error(), "required flag") {
			t.Fatalf("%v: error = %v", tt.path, err)
		}
	}
}

func TestLegacyTopLevelCommandsAreRejected(t *testing.T) {
	for _, name := range []string{"tc", "server", "user", "template", "init"} {
		cmd := NewCommand()
		cmd.SetArgs([]string{name})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		if _, err := cmd.ExecuteC(); err == nil {
			t.Fatalf("legacy top-level command %s accepted", name)
		}
	}
}
