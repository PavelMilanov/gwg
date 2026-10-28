package qos

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

const (
	rootHandle   = uint32(0x10000)  // 1:
	classHandle  = uint32(0x1fffe)  // 1:fffe, always a leaf
	leafHandle   = uint32(0x100000) // 10:
	queueLimit   = uint32(1024)
	queueFlows   = uint32(256)
	queueMemory  = uint32(262144)
	classQuantum = uint32(16384)
)

type kernelAPI interface {
	LinkByName(string) (netlink.Link, error)
	QdiscList(netlink.Link) ([]netlink.Qdisc, error)
	QdiscAdd(netlink.Qdisc) error
	QdiscDel(netlink.Qdisc) error
	ClassList(netlink.Link, uint32) ([]netlink.Class, error)
	ClassAdd(netlink.Class) error
	ClassChange(netlink.Class) error
	FilterList(netlink.Link, uint32) ([]netlink.Filter, error)
}

func rootQdisc(index int) *netlink.Htb {
	q := netlink.NewHtb(netlink.QdiscAttrs{LinkIndex: index, Handle: rootHandle, Parent: netlink.HANDLE_ROOT})
	q.Defcls = 0xfffe
	return q
}

func channelClass(index int, p Plan) *netlink.HtbClass {
	return netlink.NewHtbClass(netlink.ClassAttrs{LinkIndex: index, Handle: classHandle, Parent: rootHandle},
		netlink.HtbClassAttrs{Rate: p.RateBits, Ceil: p.RateBits, Buffer: p.BurstBytes, Cbuffer: p.BurstBytes, Quantum: classQuantum})
}

func leafQdisc(index int) *netlink.FqCodel {
	q := netlink.NewFqCodel(netlink.QdiscAttrs{LinkIndex: index, Handle: leafHandle, Parent: classHandle})
	q.Limit, q.Flows, q.MemoryLimit = queueLimit, queueFlows, queueMemory
	q.Quantum, q.Target, q.Interval = 1514, 5000, 100000
	return q
}

type snapshot struct {
	root  netlink.Qdisc
	class *netlink.HtbClass
}

// Refuse unknown children/filters instead of treating a familiar handle as ownership.
func (m *manager) inspect(link netlink.Link, owned bool) (snapshot, error) {
	qs, err := m.api.QdiscList(link)
	if err != nil {
		return snapshot{}, err
	}
	var s snapshot
	for _, q := range qs {
		if q.Attrs().Parent == netlink.HANDLE_ROOT {
			if q.Type() == "noqueue" && q.Attrs().Handle == 0 {
				continue
			}
			if s.root != nil {
				return s, fmt.Errorf("multiple root qdiscs")
			}
			s.root = q
		}
	}
	if s.root == nil {
		return s, nil
	}
	if !owned {
		return s, fmt.Errorf("root qdisc is not owned by gwu; explicit migration required")
	}
	htb, ok := s.root.(*netlink.Htb)
	if !ok || htb.Handle != rootHandle || htb.Defcls != 0xfffe || htb.Rate2Quantum != 10 {
		return s, fmt.Errorf("managed root layout changed; refusing mutation")
	}
	leafCount := 0
	for _, q := range qs {
		if q.Attrs().Parent == rootHandle || q.Attrs().Parent == classHandle {
			leaf, ok := q.(*netlink.FqCodel)
			// Kernel ns/us conversion can round default timings down by one us.
			if !ok || leaf.Handle != leafHandle || leaf.Parent != classHandle || leaf.Limit != queueLimit || leaf.Flows != queueFlows || leaf.MemoryLimit != queueMemory || leaf.Quantum != 1514 || leaf.Target < 4999 || leaf.Target > 5000 || leaf.Interval < 99999 || leaf.Interval > 100000 || leaf.ECN != 1 {
				return s, fmt.Errorf("unknown or changed child qdisc; refusing mutation: %v", q)
			}
			leafCount++
		}
	}
	if leafCount != 1 {
		return s, fmt.Errorf("managed leaf qdisc missing or duplicated")
	}
	classes, err := m.api.ClassList(link, rootHandle)
	if err != nil {
		return s, err
	}
	if len(classes) != 1 {
		return s, fmt.Errorf("unexpected class count: %d", len(classes))
	}
	c, ok := classes[0].(*netlink.HtbClass)
	if !ok {
		return s, fmt.Errorf("managed class is not HTB")
	}
	// Kernel dumps normalize a top-level class parent to HANDLE_ROOT.
	if c.Handle != classHandle || (c.Parent != rootHandle && c.Parent != netlink.HANDLE_ROOT) || c.Level != 0 || c.Rate != c.Ceil || c.Quantum != classQuantum {
		return s, fmt.Errorf("managed class layout changed: handle=%x parent=%x level=%d rate=%d ceil=%d quantum=%d", c.Handle, c.Parent, c.Level, c.Rate, c.Ceil, c.Quantum)
	}
	s.class = c
	for _, parent := range []uint32{rootHandle, classHandle} {
		filters, err := m.api.FilterList(link, parent)
		if err != nil {
			return s, err
		}
		if len(filters) != 0 {
			return s, fmt.Errorf("unexpected TC filters; refusing mutation")
		}
	}
	return s, nil
}
