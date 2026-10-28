//go:build linux

package qos

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// Only enable inside a disposable Docker container with its own network namespace.
func TestNetlinkChannelIntegration(t *testing.T) {
	if os.Getenv("GWU_TEST_NET_ADMIN") != "1" {
		t.Skip("requires isolated container and CAP_NET_ADMIN")
	}
	if _, err := os.Stat("/.dockerenv"); err != nil {
		t.Fatal("requires disposable Docker container; never use host networking")
	}
	h, err := netlink.NewHandle(unix.NETLINK_ROUTE)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	const name = "gwu-tc-test"
	if _, err := h.LinkByName(name); err == nil {
		t.Fatal("test interface already exists")
	}
	if _, err := h.LinkByName("gwu-tc-peer"); err == nil {
		t.Fatal("peer already exists")
	}
	link := &netlink.Veth{LinkAttrs: netlink.LinkAttrs{Name: name}, PeerName: "gwu-tc-peer"}
	if err := h.LinkAdd(link); err != nil {
		t.Fatal(err)
	}
	defer h.LinkDel(link)
	if err := h.LinkSetUp(link); err != nil {
		t.Fatal(err)
	}
	peer, err := h.LinkByName("gwu-tc-peer")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.LinkSetUp(peer); err != nil {
		t.Fatal(err)
	}
	// Send a small raw IPv4 frame through the real egress queue; no IP routes,
	// namespaces or shell commands are needed for the test traffic.
	socket, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW, 0x0003)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(socket)
	frame := make([]byte, 64)
	copy(frame, []byte{2, 0, 0, 0, 0, 2, 2, 0, 0, 0, 0, 1, 8, 0})
	copy(frame[14:], []byte{0x45, 0, 0, 50, 0, 0, 0, 0, 64, 17, 0, 0, 198, 18, 0, 1, 198, 18, 0, 2, 0x9c, 0x40, 0x9c, 0x41, 0, 30, 0, 0})
	m := &manager{api: h, stateDir: filepath.Join(t.TempDir(), "state")}
	c := Config{Interface: name, Rate: "75Mbit"}
	if err := m.apply(c); err != nil {
		t.Fatal("initial apply:", err)
	}
	var previousBytes uint64
	for _, rate := range []string{"75Mbit", "100Mbit", "20Mbit"} {
		c.Rate = rate
		if err := m.apply(c); err != nil {
			t.Fatal("apply", rate, err)
		}
		p, _ := c.Plan()
		state, err := m.status(name)
		if err != nil || !state.Ready || state.RateBits != p.RateBits {
			t.Fatalf("status %+v %v", state, err)
		}
		s, err := m.inspect(link, true)
		if err != nil {
			t.Fatal(err)
		}
		if s.root.(*netlink.Htb).Defcls != 0xfffe || s.class.Rate != s.class.Ceil {
			t.Fatal("invalid default class or rate")
		}
		if s.class.Statistics.Basic.Bytes < previousBytes {
			t.Fatal("class counters reset on apply")
		}
		if err := unix.Sendto(socket, frame, 0, &unix.SockaddrLinklayer{Ifindex: link.Attrs().Index, Protocol: 0x0008}); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 20; i++ {
			s, err = m.inspect(link, true)
			if err != nil {
				t.Fatal(err)
			}
			if s.class.Statistics.Basic.Bytes > previousBytes {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		if s.class.Statistics.Basic.Bytes <= previousBytes {
			t.Fatal("unclassified packet did not reach default class")
		}
		if s.root.(*netlink.Htb).DirectPkts != 0 {
			t.Fatal("packet bypassed HTB via direct queue")
		}
		previousBytes = s.class.Statistics.Basic.Bytes
	}
	// Refuse unowned ingress-independent root; even an exact matching handle is not proof.
	if err := m.remove(name); err != nil {
		t.Fatal(err)
	}
	if err := h.QdiscAdd(rootQdisc(link.Attrs().Index)); err != nil {
		t.Fatal(err)
	}
	if err := m.apply(c); err == nil {
		t.Fatal("adopted unowned root")
	}
	if err := m.remove(name); err == nil {
		t.Fatal("removed unowned root")
	}
	qs, err := h.QdiscList(link)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, q := range qs {
		if q.Attrs().Handle == rootHandle {
			found = true
		}
	}
	if !found {
		t.Fatal("foreign root disappeared")
	}
}
