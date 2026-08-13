package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCommandHierarchy(t *testing.T) {
	paths := [][]string{
		{"init"},
		{"server", "install"},
		{"server", "show"},
		{"server", "stat"},
		{"user", "add"},
		{"user", "remove"},
		{"user", "block"},
		{"user", "unblock"},
		{"user", "list"},
		{"template", "show"},
		{"template", "set"},
		{"template", "reset"},
		{"tc", "service", "up"},
		{"tc", "service", "down"},
		{"tc", "service", "restart"},
		{"tc", "service", "show"},
		{"tc", "bandwidth", "add"},
		{"tc", "bandwidth", "remove"},
		{"tc", "bandwidth", "list"},
		{"tc", "filter", "add"},
		{"tc", "filter", "remove"},
		{"tc", "filter", "list"},
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

func TestLegacyCommandIsRejected(t *testing.T) {
	rootCmd.SetArgs([]string{"add"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	if _, err := rootCmd.ExecuteC(); err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("error = %v", err)
	}
}

func TestUnsafeUserNameIsRejectedBeforeFilesystemAccess(t *testing.T) {
	rootCmd.SetArgs([]string{"user", "add", "../../root"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	if _, err := rootCmd.ExecuteC(); err == nil || !strings.Contains(err.Error(), "invalid user name") {
		t.Fatalf("error = %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	var output bytes.Buffer
	rootCmd.SetArgs([]string{"version"})
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&bytes.Buffer{})
	if _, err := rootCmd.ExecuteC(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "gwg version: "+Version) || !strings.Contains(output.String(), "go version:") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRequiredFilterFlags(t *testing.T) {
	rootCmd.SetArgs([]string{"tc", "filter", "add", "alice-limit"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	if _, err := rootCmd.ExecuteC(); err == nil || !strings.Contains(err.Error(), "required flag") {
		t.Fatalf("error = %v", err)
	}
}
