package qos

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PavelMilanov/go-wg-manager/internal/atomicfile"
	"golang.org/x/sys/unix"
)

type ownership struct {
	Version   int    `json:"version"`
	Interface string `json:"interface"`
	LinkIndex int    `json:"link_index"`
	Phase     string `json:"phase"`
}

func (m *manager) statePath(name string) string { return filepath.Join(m.stateDir, name+".json") }

func (m *manager) readOwner(name string) (*ownership, error) {
	data, err := os.ReadFile(m.statePath(name))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var owner ownership
	if err := json.Unmarshal(data, &owner); err != nil {
		return nil, err
	}
	if owner.Version != 1 || owner.Interface != name || (owner.Phase != "ready" && owner.Phase != "creating") || owner.LinkIndex <= 0 {
		return nil, fmt.Errorf("invalid TC ownership record")
	}
	return &owner, nil
}

func (m *manager) writeOwner(owner ownership) error {
	data, err := json.MarshalIndent(owner, "", "  ")
	if err != nil {
		return err
	}
	return atomicfile.Write(m.statePath(owner.Interface), data, 0600)
}

func (m *manager) lock() (func(), error) {
	if err := os.MkdirAll(m.stateDir, 0700); err != nil {
		return nil, err
	}
	info, err := os.Stat(m.stateDir)
	if err != nil {
		return nil, err
	}
	if info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("TC state directory must not be accessible to group/others")
	}
	f, err := os.OpenFile(filepath.Join(m.stateDir, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("TC operation already running: %w", err)
	}
	return func() { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN); _ = f.Close() }, nil
}
