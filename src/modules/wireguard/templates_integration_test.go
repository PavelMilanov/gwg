//go:build linux

package wireguard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Run only in a disposable Docker container with its own network namespace,
// CAP_NET_ADMIN, and GWU_TEST_NET_ADMIN=1. Never use --network=host.
func TestNativeNATLifecycleIntegration(t *testing.T) {
	if os.Getenv("GWU_TEST_NET_ADMIN") != "1" {
		t.Skip("requires explicitly enabled isolated network container")
	}
	if _, err := os.Stat("/.dockerenv"); err != nil {
		t.Fatal("integration test must run inside a disposable Docker container")
	}
	for _, command := range []string{"ip", "nft", "wg", "wg-quick", "bash"} {
		if _, err := exec.LookPath(command); err != nil {
			t.Fatal(err)
		}
	}
	run := func(command string, args ...string) string {
		t.Helper()
		out, err := exec.Command(command, args...).CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v: %s", command, args, err, out)
		}
		return string(out)
	}
	configureTestDirectories(t)
	config := validTestServerConfig()
	config.Alias, config.Eth = "gwu-test0", "gwu-uplink"
	config.ServerPrivateKey = strings.TrimSpace(run("wg", "genkey"))
	pub := exec.Command("wg", "pubkey")
	pub.Stdin = strings.NewReader(config.ServerPrivateKey + "\n")
	public, err := pub.Output()
	if err != nil {
		t.Fatal(err)
	}
	config.ServerPublicKey = strings.TrimSpace(string(public))
	table, err := natTableName(config.Alias)
	if err != nil {
		t.Fatal(err)
	}
	const foreign = "gwu_unrelated_test"
	for _, name := range []string{config.Alias, config.Eth} {
		if exec.Command("ip", "link", "show", "dev", name).Run() == nil {
			t.Fatalf("interface %s already exists", name)
		}
	}
	for _, name := range []string{table, foreign} {
		if exec.Command("nft", "list", "table", "ip", name).Run() == nil {
			t.Fatalf("table %s already exists", name)
		}
	}
	run("ip", "link", "add", config.Eth, "type", "dummy")
	t.Cleanup(func() { _ = exec.Command("ip", "link", "delete", config.Eth).Run() })
	run("nft", "create", "table", "ip", foreign)
	t.Cleanup(func() { _ = exec.Command("nft", "delete", "table", "ip", foreign).Run() })
	foreignBefore := run("nft", "-s", "list", "table", "ip", foreign)
	if err := writeServerConfig(config, config.Alias); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(serverDir, config.Alias+".conf")
	t.Cleanup(func() {
		_ = exec.Command("ip", "link", "delete", config.Alias).Run()
		_ = exec.Command("nft", "delete", "table", "ip", table).Run()
	})
	for i := 0; i < 2; i++ {
		run("wg-quick", "up", path)
		state := run("nft", "list", "table", "ip", table)
		if strings.Count(state, "masquerade") != 1 || !strings.Contains(state, `oifname "gwu-uplink"`) {
			t.Fatal("unexpected NAT rules")
		}
		if run("nft", "-s", "list", "table", "ip", foreign) != foreignBefore {
			t.Fatal("foreign table changed")
		}
		run("wg-quick", "down", path)
		if exec.Command("nft", "list", "table", "ip", table).Run() == nil {
			t.Fatal("NAT table remains after down")
		}
	}
	// A name collision must fail without adopting, flushing or changing that table.
	run("nft", "create", "table", "ip", table)
	before := []byte(run("nft", "-s", "list", "table", "ip", table))
	if exec.Command("wg-quick", "up", path).Run() == nil {
		t.Fatal("accepted pre-existing NAT table")
	}
	if !bytes.Equal(before, []byte(run("nft", "-s", "list", "table", "ip", table))) {
		t.Fatal("conflicting table changed")
	}
	if run("nft", "-s", "list", "table", "ip", foreign) != foreignBefore {
		t.Fatal("foreign table changed")
	}
	if exec.Command("ip", "link", "show", "dev", config.Alias).Run() == nil {
		t.Fatal("failed up left WireGuard interface")
	}
}
