package tc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	output []byte
	err    error
	calls  []string
}

func (f *fakeRunner) Output(name string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, strings.Join(append([]string{name}, args...), " "))
	return f.output, f.err
}

func (f *fakeRunner) OutputWithInput(_ []byte, name string, args ...string) ([]byte, error) {
	return f.Output(name, args...)
}

func (f *fakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, strings.Join(append([]string{name}, args...), " "))
	return f.err
}

func TestValidateRateRejectsShellInput(t *testing.T) {
	for _, value := range []string{"", "0Mbit", "10 Mbit", "10Mbit;id", "$(id)"} {
		if err := validateRate("rate", value); err == nil {
			t.Errorf("validateRate(%q) succeeded", value)
		}
	}
	for _, value := range []string{"1bit", "50Mbit", "2gbit"} {
		if err := validateRate("rate", value); err != nil {
			t.Errorf("validateRate(%q): %v", value, err)
		}
	}
}

func TestNextClassIDReusesGap(t *testing.T) {
	classes := []TcClass{{Class: "2"}, {Class: "4"}}
	if got := nextClassID(classes); got != "3" {
		t.Fatalf("nextClassID = %s, want 3", got)
	}
}

func TestValidateConfigRejectsDanglingFilter(t *testing.T) {
	config := TcConfig{
		Intf:      "vpn7",
		Speed:     "10Mbit",
		FullSpeed: "20Mbit",
		Filters:   []TcFilter{{Description: "alice", UserIp: "10.0.0.2/32", Class: "2"}},
	}
	if err := validateConfig(config); err == nil || !strings.Contains(err.Error(), "missing class") {
		t.Fatalf("error = %v", err)
	}
}

func TestGenerateUsesConfiguredInterface(t *testing.T) {
	oldDir := tcDir
	tcDir = t.TempDir()
	t.Cleanup(func() { tcDir = oldDir })
	config := TcConfig{Intf: "vpn7", Speed: "10Mbit", FullSpeed: "20Mbit"}
	if err := config.generate(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(tcDir, "tc.sh"))
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	if !strings.Contains(script, "dev vpn7") || strings.Contains(script, "wg0") || strings.Contains(script, "sudo tc") {
		t.Fatalf("unexpected generated script:\n%s", script)
	}
}

func TestBandwidthLifecycle(t *testing.T) {
	oldDir := tcDir
	tcDir = t.TempDir()
	t.Cleanup(func() { tcDir = oldDir })

	if err := AddBandwidth("regular", "2Mbit", "3Mbit"); err != nil {
		t.Fatal(err)
	}
	if err := AddBandwidth("premium", "10Mbit", "20Mbit"); err != nil {
		t.Fatal(err)
	}
	classes, err := readClassFile()
	if err != nil {
		t.Fatal(err)
	}
	if len(classes) != 2 || classes[0].Class != "2" || classes[1].Class != "3" {
		t.Fatalf("classes = %#v", classes)
	}
	if err := writeJSONFile("filters", []TcFilter{{Description: "alice", UserIp: "10.0.0.2/32", Class: "2"}}); err != nil {
		t.Fatal(err)
	}
	if err := RemoveBandwidth("2"); err == nil || !strings.Contains(err.Error(), "used by filter") {
		t.Fatalf("error = %v", err)
	}
	if err := RemoveFilter("alice"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveBandwidth("2"); err != nil {
		t.Fatal(err)
	}
	classes, err = readClassFile()
	if err != nil {
		t.Fatal(err)
	}
	if len(classes) != 1 || classes[0].Class != "3" {
		t.Fatalf("classes after removal = %#v", classes)
	}
}

func TestCreateServiceUsesConfiguredWireGuardUnit(t *testing.T) {
	oldTcDir, oldServiceDir := tcDir, serviceDir
	tcDir, serviceDir = t.TempDir(), t.TempDir()
	t.Cleanup(func() { tcDir, serviceDir = oldTcDir, oldServiceDir })
	fake := &fakeRunner{}
	previous := commandRunner
	commandRunner = fake
	t.Cleanup(func() { commandRunner = previous })

	config := TcConfig{Intf: "vpn7", Speed: "10Mbit", FullSpeed: "20Mbit"}
	if err := config.createService(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(tcDir, "tc.service"))
	if err != nil {
		t.Fatal(err)
	}
	service := string(data)
	if !strings.Contains(service, "After=wg-quick@vpn7.service") || strings.Contains(service, "wg0") {
		t.Fatalf("unexpected service:\n%s", service)
	}
	if len(fake.calls) != 4 {
		t.Fatalf("calls = %v", fake.calls)
	}
}

func TestServiceStateTrimsNewline(t *testing.T) {
	fake := &fakeRunner{output: []byte("enabled\n")}
	previous := commandRunner
	commandRunner = fake
	t.Cleanup(func() { commandRunner = previous })
	state, err := serviceState()
	if err != nil {
		t.Fatal(err)
	}
	if state != "enabled" {
		t.Fatalf("state = %q", state)
	}
}

func TestDownUsesConfiguredInterfaceWithoutShell(t *testing.T) {
	fake := &fakeRunner{}
	previous := commandRunner
	commandRunner = fake
	t.Cleanup(func() { commandRunner = previous })
	if err := (&TcConfig{Intf: "vpn7"}).down(); err != nil {
		t.Fatal(err)
	}
	want := "sudo tc qdisc del dev vpn7 root"
	if len(fake.calls) != 1 || fake.calls[0] != want {
		t.Fatalf("calls = %v, want %q", fake.calls, want)
	}
}

func TestServiceStateReturnsDisabledOutputWithError(t *testing.T) {
	fake := &fakeRunner{output: []byte("disabled\n"), err: errors.New("exit status 1")}
	previous := commandRunner
	commandRunner = fake
	t.Cleanup(func() { commandRunner = previous })
	state, err := serviceState()
	if state != "disabled" || err == nil {
		t.Fatalf("state, error = %q, %v", state, err)
	}
}
