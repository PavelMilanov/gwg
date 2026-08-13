package server

import (
	"encoding/base64"
	"fmt"
	"net/netip"
	"path/filepath"
	"regexp"
	"strings"
)

var userNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
var interfaceNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.+=-]{0,14}$`)

func validateUserName(name string) error {
	if !userNamePattern.MatchString(name) || name == "." || name == ".." {
		return fmt.Errorf("invalid user name %q: use 1-64 letters, digits, dots, dashes or underscores", name)
	}
	return nil
}

func validateInterfaceName(name string) error {
	if !interfaceNamePattern.MatchString(name) || name == "." || name == ".." {
		return fmt.Errorf("invalid WireGuard interface name %q", name)
	}
	return nil
}

func parseServerPrefix(value string) (netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(value)
	if err != nil || !prefix.Addr().Is4() {
		return netip.Prefix{}, fmt.Errorf("invalid IPv4 prefix %q", value)
	}
	if prefix.Bits() > 30 {
		return netip.Prefix{}, fmt.Errorf("IPv4 prefix %q has no client addresses", value)
	}
	if prefix.Addr() == prefix.Masked().Addr() || prefix.Addr() == lastAddress(prefix) {
		return netip.Prefix{}, fmt.Errorf("IPv4 address %q is reserved by its network", prefix.Addr())
	}
	return prefix, nil
}

func validateServerConfig(config WgServerConfig) error {
	if err := validateInterfaceName(config.Alias); err != nil {
		return err
	}
	if err := validateInterfaceName(config.Eth); err != nil {
		return fmt.Errorf("invalid external interface: %w", err)
	}
	if _, err := parseServerPrefix(config.LocalAddress); err != nil {
		return err
	}
	if ip, err := netip.ParseAddr(config.PublicAddress); err != nil || !ip.Is4() {
		return fmt.Errorf("invalid server endpoint address %q", config.PublicAddress)
	}
	if config.ListenPort < 1 || config.ListenPort > 65535 {
		return fmt.Errorf("invalid WireGuard port %d", config.ListenPort)
	}
	if err := validateKey("server private key", config.ServerPrivateKey); err != nil {
		return err
	}
	if err := validateKey("server public key", config.ServerPublicKey); err != nil {
		return err
	}
	if err := validateSingleLine("DNS resolver", config.DnsResolv); err != nil {
		return err
	}
	names := make(map[string]struct{}, len(config.Users))
	addresses := make(map[string]struct{}, len(config.Users))
	publicKeys := make(map[string]struct{}, len(config.Users))
	for _, user := range config.Users {
		if err := validateUserConfig(user); err != nil {
			return fmt.Errorf("invalid user %q: %w", user.Name, err)
		}
		if _, exists := names[user.Name]; exists {
			return fmt.Errorf("duplicate user name %q", user.Name)
		}
		if _, exists := addresses[user.ClientLocalAddress]; exists {
			return fmt.Errorf("duplicate client address %q", user.ClientLocalAddress)
		}
		if _, exists := publicKeys[user.ClientPublicKey]; exists {
			return fmt.Errorf("duplicate client public key for user %q", user.Name)
		}
		names[user.Name] = struct{}{}
		addresses[user.ClientLocalAddress] = struct{}{}
		publicKeys[user.ClientPublicKey] = struct{}{}
	}
	return nil
}

func validateUserConfig(config UserConfig) error {
	if err := validateUserName(config.Name); err != nil {
		return err
	}
	clientPrefix, err := netip.ParsePrefix(config.ClientLocalAddress)
	if err != nil || !clientPrefix.Addr().Is4() || clientPrefix.Bits() != 32 {
		return fmt.Errorf("invalid client IPv4 address %q", config.ClientLocalAddress)
	}
	if ip, err := netip.ParseAddr(config.ServerIp); err != nil || !ip.Is4() {
		return fmt.Errorf("invalid server IPv4 address %q", config.ServerIp)
	}
	if config.ServerPort < 1 || config.ServerPort > 65535 {
		return fmt.Errorf("invalid WireGuard port %d", config.ServerPort)
	}
	if err := validateKey("client private key", config.ClientPrivateKey); err != nil {
		return err
	}
	if err := validateKey("client public key", config.ClientPublicKey); err != nil {
		return err
	}
	if err := validateKey("server public key", config.ServerPublicKey); err != nil {
		return err
	}
	if err := validateSingleLine("DNS resolver", config.DnsResolv); err != nil {
		return err
	}
	if config.Status != "" && config.Status != "active" {
		return fmt.Errorf("invalid user status %q", config.Status)
	}
	return nil
}

func validateKey(name, value string) error {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return fmt.Errorf("invalid %s", name)
	}
	return nil
}

func validateSingleLine(name, value string) error {
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("invalid %s", name)
	}
	return nil
}

func configPath(dir, name, extension string) string {
	return filepath.Join(dir, name+extension)
}
