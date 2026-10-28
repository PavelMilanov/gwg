package qos

import (
	"errors"
	"fmt"
	"os"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type manager struct {
	api      kernelAPI
	stateDir string
}

type Status struct {
	Interface string `json:"interface"`
	Owned     bool   `json:"owned"`
	Ready     bool   `json:"ready"`
	RateBits  uint64 `json:"rate_bits"`
	Detail    string `json:"detail,omitempty"`
}

func withKernel(fn func(*manager) error) error {
	h, err := netlink.NewHandle(unix.NETLINK_ROUTE)
	if err != nil {
		return err
	}
	defer h.Close()
	return fn(&manager{api: h, stateDir: DefaultStateDir})
}

func Apply(c Config) error {
	return withKernel(func(m *manager) error { return m.apply(c) })
}

func Remove(name string) error {
	return withKernel(func(m *manager) error { return m.remove(name) })
}

func Inspect(name string) (Status, error) {
	var result Status
	err := withKernel(func(m *manager) error { var err error; result, err = m.status(name); return err })
	return result, err
}

func (m *manager) apply(c Config) error {
	p, err := c.Plan()
	if err != nil {
		return err
	}
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	link, err := m.api.LinkByName(c.Interface)
	if err != nil {
		return err
	}
	owner, err := m.readOwner(c.Interface)
	if err != nil {
		return err
	}
	owned := owner != nil && owner.LinkIndex == link.Attrs().Index
	s, err := m.inspect(link, owned)
	if err != nil {
		return err
	}
	if s.root != nil {
		if owner.Phase != "ready" {
			return fmt.Errorf("unfinished TC operation; inspect and recover explicitly")
		}
		desired := channelClass(link.Attrs().Index, p)
		if s.class.Rate == desired.Rate && s.class.Buffer == desired.Buffer && s.class.Cbuffer == desired.Cbuffer {
			return nil
		}
		// Only the existing class changes. Root, FQ-CoDel and other hooks remain.
		if err := m.api.ClassChange(desired); err != nil {
			return fmt.Errorf("change shared class: %w", err)
		}
		verified, err := m.inspect(link, true)
		if err != nil {
			return err
		}
		if verified.class == nil || verified.class.Rate != p.RateBits/8 {
			return fmt.Errorf("kernel did not confirm requested rate")
		}
		return nil
	}
	record := ownership{Version: 1, Interface: c.Interface, LinkIndex: link.Attrs().Index, Phase: "creating"}
	if err := m.writeOwner(record); err != nil {
		return err
	}
	root := rootQdisc(link.Attrs().Index)
	if err := m.api.QdiscAdd(root); err != nil {
		// No root was created by this operation: never delete a competing qdisc.
		return errors.Join(err, os.Remove(m.statePath(c.Interface)))
	}
	rollback := func(cause error) error {
		if err := m.api.QdiscDel(root); err != nil {
			return errors.Join(cause, fmt.Errorf("rollback failed; ownership retained: %w", err))
		}
		return errors.Join(cause, os.Remove(m.statePath(c.Interface)))
	}
	if err := m.api.ClassAdd(channelClass(link.Attrs().Index, p)); err != nil {
		return rollback(err)
	}
	if err := m.api.QdiscAdd(leafQdisc(link.Attrs().Index)); err != nil {
		return rollback(err)
	}
	verified, err := m.inspect(link, true)
	if err != nil {
		return rollback(err)
	}
	if verified.class == nil || verified.class.Rate != p.RateBits/8 {
		return rollback(fmt.Errorf("kernel did not confirm initial rate"))
	}
	record.Phase = "ready"
	if err := m.writeOwner(record); err != nil {
		return rollback(err)
	}
	return nil
}

func (m *manager) remove(name string) error {
	if err := validateInterface(name); err != nil {
		return err
	}
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	owner, err := m.readOwner(name)
	if err != nil {
		return err
	}
	if owner == nil {
		return fmt.Errorf("no gwu ownership record; refusing deletion")
	}
	link, err := m.api.LinkByName(name)
	if err != nil {
		return err
	}
	s, err := m.inspect(link, owner.LinkIndex == link.Attrs().Index)
	if err != nil {
		return err
	}
	if s.root != nil {
		if owner.Phase != "ready" {
			return fmt.Errorf("unfinished TC operation; explicit recovery required")
		}
		if err := m.api.QdiscDel(s.root); err != nil {
			return err
		}
	}
	return os.Remove(m.statePath(name))
}

func (m *manager) status(name string) (Status, error) {
	result := Status{Interface: name}
	if err := validateInterface(name); err != nil {
		return result, err
	}
	link, err := m.api.LinkByName(name)
	if err != nil {
		return result, err
	}
	owner, err := m.readOwner(name)
	if err != nil {
		return result, err
	}
	result.Owned = owner != nil && owner.LinkIndex == link.Attrs().Index
	s, err := m.inspect(link, result.Owned)
	if err != nil {
		result.Detail = err.Error()
		return result, nil
	}
	if s.class != nil {
		result.RateBits = s.class.Rate * 8
		result.Ready = owner.Phase == "ready"
	}
	if s.root == nil {
		result.Detail = "no managed root qdisc"
	}
	return result, nil
}
