package qos

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/vishvananda/netlink"
)

type fakeKernel struct {
	link    netlink.Link
	qdiscs  []netlink.Qdisc
	classes []netlink.Class
	filters []netlink.Filter
	ops     []string
	fail    string
}

func (f *fakeKernel) LinkByName(string) (netlink.Link, error)                 { return f.link, nil }
func (f *fakeKernel) QdiscList(netlink.Link) ([]netlink.Qdisc, error)         { return f.qdiscs, nil }
func (f *fakeKernel) ClassList(netlink.Link, uint32) ([]netlink.Class, error) { return f.classes, nil }
func (f *fakeKernel) FilterList(netlink.Link, uint32) ([]netlink.Filter, error) {
	return f.filters, nil
}
func (f *fakeKernel) op(name string) error {
	f.ops = append(f.ops, name)
	if f.fail == name {
		return errors.New("injected " + name)
	}
	return nil
}
func (f *fakeKernel) QdiscAdd(q netlink.Qdisc) error {
	if err := f.op("add " + q.Type()); err != nil {
		return err
	}
	f.qdiscs = append(f.qdiscs, q)
	return nil
}
func (f *fakeKernel) QdiscDel(netlink.Qdisc) error {
	if err := f.op("delete root"); err != nil {
		return err
	}
	f.qdiscs = nil
	f.classes = nil
	return nil
}
func (f *fakeKernel) ClassAdd(c netlink.Class) error {
	if err := f.op("add class"); err != nil {
		return err
	}
	f.classes = []netlink.Class{c}
	return nil
}
func (f *fakeKernel) ClassChange(c netlink.Class) error {
	if err := f.op("change class"); err != nil {
		return err
	}
	f.classes = []netlink.Class{c}
	return nil
}
func testManager(t *testing.T) (*manager, *fakeKernel) {
	t.Helper()
	f := &fakeKernel{link: &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "vpn7", Index: 7}}}
	return &manager{api: f, stateDir: filepath.Join(t.TempDir(), "state")}, f
}
func TestPlan(t *testing.T) {
	for _, tt := range []struct {
		rate  string
		bits  uint64
		burst uint32
	}{{"8Kbit", 8000, 16384}, {"75Mbit", 75000000, 65536}, {"950Mbit", 950000000, 524288}, {"2Gbit", 2000000000, 1048576}} {
		p, err := (Config{Interface: "vpn7", Rate: tt.rate}).Plan()
		if err != nil || p.RateBits != tt.bits || p.BurstBytes != tt.burst {
			t.Fatalf("%s: %+v %v", tt.rate, p, err)
		}
	}
	for _, rate := range []string{"", "0Mbit", "1bit", "8001bit", "999999999999999999999Gbit", "101Gbit", "10 Mbit", "10Mbit;id", "$(id)"} {
		if _, err := (Config{Interface: "vpn7", Rate: rate}).Plan(); err == nil {
			t.Fatalf("accepted %q", rate)
		}
	}
	if _, err := (Config{Interface: "../x", Rate: "10Mbit"}).Plan(); err == nil {
		t.Fatal("unsafe interface")
	}
}
func TestReadConfigRejectsLegacyAndTrailingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "channel.json")
	for _, data := range []string{`{"Intf":"wg0","Speed":"75Mbit","FullSpeed":"950Mbit"}`, `{"interface":"wg0","rate":"75Mbit","classes":[]}`, `{"interface":"wg0","rate":"75Mbit"} {}`} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadConfig(path); err == nil {
			t.Fatalf("accepted %s", data)
		}
	}
}
func TestApplyUpdatesOnlyClassAndIsIdempotent(t *testing.T) {
	m, f := testManager(t)
	c := Config{Interface: "vpn7", Rate: "75Mbit"}
	if err := m.apply(c); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.ops, []string{"add htb", "add class", "add fq_codel"}) {
		t.Fatal(f.ops)
	}
	root, leaf := f.qdiscs[0], f.qdiscs[1]
	f.ops = nil
	if err := m.apply(c); err != nil {
		t.Fatal(err)
	}
	if len(f.ops) != 0 {
		t.Fatal("idempotent apply mutated kernel", f.ops)
	}
	c.Rate = "100Mbit"
	if err := m.apply(c); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.ops, []string{"change class"}) || f.qdiscs[0] != root || f.qdiscs[1] != leaf {
		t.Fatal("root or leaf recreated", f.ops)
	}
	s, err := m.status(c.Interface)
	if err != nil || !s.Ready || s.RateBits != 100000000 {
		t.Fatalf("%+v %v", s, err)
	}
	if err := m.remove(c.Interface); err != nil {
		t.Fatal(err)
	}
	if len(f.qdiscs) != 0 {
		t.Fatal("root remains")
	}
}
func TestRefuseUnownedRootEvenWithMatchingHandles(t *testing.T) {
	m, f := testManager(t)
	f.qdiscs = []netlink.Qdisc{rootQdisc(7), leafQdisc(7)}
	if err := m.apply(Config{Interface: "vpn7", Rate: "75Mbit"}); err == nil {
		t.Fatal("adopted unowned root")
	}
	if err := m.remove("vpn7"); err == nil {
		t.Fatal("deleted unowned root")
	}
	if len(f.ops) != 0 {
		t.Fatal(f.ops)
	}
}

func TestKernelDumpNormalizationIsAccepted(t *testing.T) {
	m, f := testManager(t)
	c := Config{Interface: "vpn7", Rate: "75Mbit"}
	if err := m.apply(c); err != nil {
		t.Fatal(err)
	}
	f.classes[0].Attrs().Parent = netlink.HANDLE_ROOT
	leaf := f.qdiscs[1].(*netlink.FqCodel)
	leaf.Target, leaf.Interval = 4999, 99999
	f.ops = nil
	if err := m.apply(c); err != nil {
		t.Fatal(err)
	}
	if len(f.ops) != 0 {
		t.Fatal("normalized kernel dump triggered mutation", f.ops)
	}
}
func TestRefuseExtraObjectsAndStaleOwner(t *testing.T) {
	for _, mode := range []string{"filter", "class", "leaf", "index"} {
		t.Run(mode, func(t *testing.T) {
			m, f := testManager(t)
			c := Config{Interface: "vpn7", Rate: "75Mbit"}
			if err := m.apply(c); err != nil {
				t.Fatal(err)
			}
			f.ops = nil
			switch mode {
			case "filter":
				f.filters = []netlink.Filter{&netlink.U32{}}
			case "class":
				f.classes = append(f.classes, f.classes[0])
			case "leaf":
				f.qdiscs[1].Attrs().Handle = 0x200000
			case "index":
				f.link.Attrs().Index = 8
			}
			if err := m.apply(c); err == nil {
				t.Fatal("accepted changed layout")
			}
			if err := m.remove(c.Interface); err == nil {
				t.Fatal("deleted changed layout")
			}
			if len(f.ops) != 0 {
				t.Fatal(f.ops)
			}
		})
	}
}
func TestCreationFailureRollsBackOnlyNewRoot(t *testing.T) {
	for _, failure := range []string{"add htb", "add class", "add fq_codel"} {
		t.Run(failure, func(t *testing.T) {
			m, f := testManager(t)
			f.fail = failure
			if err := m.apply(Config{Interface: "vpn7", Rate: "75Mbit"}); err == nil {
				t.Fatal("expected failure")
			}
			if len(f.qdiscs) != 0 {
				t.Fatal("partial root remains")
			}
			if _, err := os.Stat(m.statePath("vpn7")); !os.IsNotExist(err) {
				t.Fatal("ownership remains", err)
			}
			if failure == "add htb" && len(f.ops) != 1 {
				t.Fatal("deleted competing root", f.ops)
			}
		})
	}
}
func TestFailedClassChangeKeepsExistingObjects(t *testing.T) {
	m, f := testManager(t)
	c := Config{Interface: "vpn7", Rate: "75Mbit"}
	if err := m.apply(c); err != nil {
		t.Fatal(err)
	}
	f.ops = nil
	f.fail = "change class"
	c.Rate = "100Mbit"
	if err := m.apply(c); err == nil {
		t.Fatal("expected failure")
	}
	if !reflect.DeepEqual(f.ops, []string{"change class"}) || len(f.qdiscs) != 2 {
		t.Fatal(f.ops)
	}
}
func TestLockBlocksConcurrentMutation(t *testing.T) {
	m, f := testManager(t)
	unlock, err := m.lock()
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if err := m.apply(Config{Interface: "vpn7", Rate: "75Mbit"}); err == nil {
		t.Fatal("ignored lock")
	}
	if len(f.ops) != 0 {
		t.Fatal(f.ops)
	}
}
func TestServiceTemplateDoesNotInvokeShellOrDeleteRoot(t *testing.T) {
	for _, word := range []string{"tc.sh", "ExecStartPre", "ExecStop", "qdisc del"} {
		if strings.Contains(ServiceTemplate, word) {
			t.Fatal(word)
		}
	}
	unit, err := RenderService(Config{Interface: "vpn7", Rate: "75Mbit"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(unit), "After=wg-quick@vpn7.service") || !strings.Contains(string(unit), "ExecReload=/usr/bin/gwu qos apply") || strings.Contains(string(unit), "wg0") {
		t.Fatal("unexpected unit")
	}
}
