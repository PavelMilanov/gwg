// Package qos implements a single shared egress channel through Linux Netlink.
package qos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PavelMilanov/go-wg-manager/internal/layout"
)

const DefaultConfigPath = layout.QoSDir + "/channel.json"
const DefaultStateDir = layout.QoSDir + "/state"

type Config struct {
	Interface string `json:"interface"`
	Rate      string `json:"rate"`
}

type Plan struct {
	Interface  string `json:"interface"`
	RateBits   uint64 `json:"rate_bits"`
	BurstBytes uint32 `json:"burst_bytes"`
}

var interfacePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.+=-]{0,14}$`)
var ratePattern = regexp.MustCompile(`^([1-9][0-9]*)(bit|kbit|mbit|gbit)$`)

func validateInterface(name string) error {
	if !interfacePattern.MatchString(name) {
		return fmt.Errorf("invalid interface name %q", name)
	}
	return nil
}

func (c Config) Plan() (Plan, error) {
	if err := validateInterface(c.Interface); err != nil {
		return Plan{}, err
	}
	m := ratePattern.FindStringSubmatch(strings.ToLower(c.Rate))
	if m == nil {
		return Plan{}, fmt.Errorf("invalid rate %q: use e.g. 75Mbit", c.Rate)
	}
	n, err := strconv.ParseUint(m[1], 10, 64)
	scale := map[string]uint64{"bit": 1, "kbit": 1000, "mbit": 1000000, "gbit": 1000000000}[m[2]]
	if err != nil || n > 100000000000/scale {
		return Plan{}, fmt.Errorf("rate exceeds prototype limit 100Gbit")
	}
	bits := n * scale
	if bits < 8000 || bits%8 != 0 {
		return Plan{}, fmt.Errorf("rate must be at least 8Kbit and a multiple of 8bit")
	}
	needed := (bits + 1999) / 2000 // four milliseconds of tokens in bytes
	burst := uint32(16384)
	for uint64(burst) < needed {
		burst <<= 1
	}
	return Plan{Interface: c.Interface, RateBits: bits, BurstBytes: burst}, nil
}

func ReadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open channel config: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 1024*1024+1))
	if err != nil {
		return Config{}, err
	}
	if len(data) > 1024*1024 {
		return Config{}, fmt.Errorf("channel config exceeds 1 MiB")
	}
	return decodeConfig(data)
}

func decodeConfig(data []byte) (Config, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var c Config
	if err := decoder.Decode(&c); err != nil {
		return Config{}, fmt.Errorf("decode channel config: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Config{}, fmt.Errorf("expected exactly one channel config object")
	}
	_, err := c.Plan()
	return c, err
}
