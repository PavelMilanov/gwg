package wireguard

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
)

func setClientIP(server WgServerConfig, configs []UserConfig) (string, error) {
	prefix, err := parseServerPrefix(server.LocalAddress)
	if err != nil {
		return "", err
	}

	used := make(map[netip.Addr]struct{}, len(configs)+1)
	used[prefix.Addr()] = struct{}{}
	for _, config := range configs {
		clientPrefix, err := netip.ParsePrefix(config.ClientLocalAddress)
		if err != nil || !clientPrefix.Addr().Is4() {
			return "", fmt.Errorf("user %q has invalid address %q", config.Name, config.ClientLocalAddress)
		}
		if !prefix.Contains(clientPrefix.Addr()) {
			return "", fmt.Errorf("user %q address %q is outside %q", config.Name, config.ClientLocalAddress, prefix)
		}
		used[clientPrefix.Addr()] = struct{}{}
	}

	network := prefix.Masked().Addr()
	last := lastAddress(prefix)
	for candidate := network.Next(); candidate.IsValid() && candidate.Less(last); candidate = candidate.Next() {
		if _, exists := used[candidate]; !exists {
			return netip.PrefixFrom(candidate, 32).String(), nil
		}
	}
	return "", fmt.Errorf("IPv4 prefix %q has no free client addresses", prefix)
}

func lastAddress(prefix netip.Prefix) netip.Addr {
	bytes := prefix.Masked().Addr().As4()
	value := uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
	value |= ^uint32(0) >> uint(prefix.Bits())
	return netip.AddrFrom4([4]byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)})
}

func setServerParams() (string, string, error) {
	out, err := commandRunner.Output("ip", "route", "show", "default")
	if err != nil {
		return "", "", err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return "", "", fmt.Errorf("default route not found")
	}

	fields := strings.Fields(lines[0])
	var interfaceName, sourceIP string
	for i := 0; i+1 < len(fields); i++ {
		switch fields[i] {
		case "dev":
			interfaceName = fields[i+1]
		case "src":
			sourceIP = fields[i+1]
		}
	}
	if interfaceName == "" {
		return "", "", fmt.Errorf("default route has no interface: %q", lines[0])
	}
	if err := validateInterfaceName(interfaceName); err != nil {
		return "", "", fmt.Errorf("invalid default-route interface: %w", err)
	}
	if sourceIP != "" {
		if ip, err := netip.ParseAddr(sourceIP); err == nil && ip.Is4() {
			return ip.String(), interfaceName, nil
		}
	}

	interf, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", "", fmt.Errorf("find interface %s: %w", interfaceName, err)
	}
	addrs, err := interf.Addrs()
	if err != nil {
		return "", "", fmt.Errorf("read addresses for interface %s: %w", interfaceName, err)
	}
	for _, addr := range addrs {
		prefix, err := netip.ParsePrefix(addr.String())
		if err == nil && prefix.Addr().Is4() {
			return prefix.Addr().String(), interfaceName, nil
		}
	}
	return "", "", fmt.Errorf("interface %s has no IPv4 address", interfaceName)
}
