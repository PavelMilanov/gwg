package server

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/PavelMilanov/go-wg-manager/internal/command"
)

var commandRunner command.Runner = command.ExecRunner{}

func initSystem() error {
	if _, err := os.Stat("/etc/os-release"); err != nil {
		return fmt.Errorf("operating system is not supported: %w", err)
	}
	return nil
}

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

func generateKeys() (string, string, error) {
	privateOutput, err := commandRunner.Output("wg", "genkey")
	if err != nil {
		return "", "", fmt.Errorf("generate WireGuard private key: %w", err)
	}
	privateKey := strings.TrimSpace(string(privateOutput))
	if err := validateKey("generated private key", privateKey); err != nil {
		return "", "", err
	}

	publicOutput, err := commandRunner.OutputWithInput([]byte(privateKey+"\n"), "wg", "pubkey")
	if err != nil {
		return "", "", fmt.Errorf("generate WireGuard public key: %w", err)
	}
	publicKey := strings.TrimSpace(string(publicOutput))
	if err := validateKey("generated public key", publicKey); err != nil {
		return "", "", err
	}
	return privateKey, publicKey, nil
}

func ShowPeers() error {
	out, err := commandRunner.Output("sudo", "wg", "show")
	if err != nil {
		return err
	}
	fmt.Print(string(out))
	return nil
}

func commandServer(action string, alias string) error {
	if err := validateInterfaceName(alias); err != nil {
		return err
	}
	switch action {
	case "enable", "start", "restart":
	default:
		return fmt.Errorf("unsupported server action %q", action)
	}
	service := "wg-quick@" + alias + ".service"
	if err := commandRunner.Run("sudo", "systemctl", action, service); err != nil {
		return fmt.Errorf("%s %s: %w", action, service, err)
	}
	return nil
}

func parseDump(data []byte, configs []UserConfig) ([]WireguardDump, error) {
	names := make(map[string]string, len(configs))
	for _, config := range configs {
		names[config.ClientPublicKey] = config.Name
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) <= 1 || (len(lines) == 1 && lines[0] == "") {
		return nil, nil
	}
	pool := make([]WireguardDump, 0, len(lines)-1)
	for lineNumber, line := range lines[1:] {
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			return nil, fmt.Errorf("invalid wg dump line %d: expected at least 7 fields, got %d", lineNumber+2, len(fields))
		}
		rx, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse received bytes on dump line %d: %w", lineNumber+2, err)
		}
		tx, err := strconv.ParseUint(fields[6], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse sent bytes on dump line %d: %w", lineNumber+2, err)
		}
		pool = append(pool, WireguardDump{user: names[fields[0]], ip: fields[3], rateRx: rx, rateTx: tx})
	}
	return pool, nil
}
