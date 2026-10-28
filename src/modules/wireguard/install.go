package wireguard

import (
	"errors"
	"fmt"
	"os"

	"github.com/PavelMilanov/go-wg-manager/internal/atomicfile"
)

func InstallServer(alias, network string, port int) error {
	if err := validateInterfaceName(alias); err != nil {
		return err
	}
	if _, err := parseServerPrefix(network); err != nil {
		return err
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid WireGuard port %d", port)
	}
	// New module paths must never make an existing unmanaged interface look new.
	if _, err := os.Lstat(configPath(serverDir, alias, ".conf")); err == nil {
		return fmt.Errorf("WireGuard configuration %s already exists; explicit migration required", alias)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, dir := range []string{serverDir, managerDir, userConfigDir, usersDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	if err := ensureConfigTemplates(); err != nil {
		return err
	}
	metadataPath := configPath(managerDir, alias, ".json")
	if _, err := os.Stat(metadataPath); err == nil {
		return fmt.Errorf("server %q is already configured", alias)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect %s: %w", metadataPath, err)
	}

	privateKey, publicKey, err := generateKeys()
	if err != nil {
		return err
	}
	publicAddress, externalInterface, err := setServerParams()
	if err != nil {
		return err
	}
	config := WgServerConfig{
		ServerPrivateKey: privateKey,
		ServerPublicKey:  publicKey,
		LocalAddress:     network,
		PublicAddress:    publicAddress,
		ListenPort:       port,
		Eth:              externalInterface,
		DnsResolv:        "8.8.8.8",
		Alias:            alias,
	}
	if err := writeServerConfig(config, alias); err != nil {
		return err
	}
	if err := config.createServerConfigFile(); err != nil {
		return err
	}
	if err := commandServer("enable", alias); err != nil {
		return err
	}
	if err := commandServer("start", alias); err != nil {
		return err
	}
	fmt.Println("Server started")
	return nil
}

func ConfigureSystem() error {
	if _, err := os.Stat("/etc/os-release"); err != nil {
		return fmt.Errorf("operating system is not supported: %w", err)
	}
	if os.Geteuid() != 0 {
		return fmt.Errorf("initialization requires root privileges; run sudo gwu wireguard init")
	}
	if err := prepareSystem(); err != nil {
		return err
	}
	return InstallServer("wg0", "10.0.0.1/24", 51830)
}

func prepareSystem() error {
	for _, dir := range []string{serverDir, managerDir, userConfigDir, usersDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
		if err := os.Chmod(dir, 0700); err != nil {
			return fmt.Errorf("set permissions on directory %s: %w", dir, err)
		}
	}
	if err := ensureConfigTemplates(); err != nil {
		return err
	}
	output, err := gatewayTemplate.render(WgServerConfig{})
	if err != nil {
		return err
	}
	if err := writeGenerated("gateway.conf", output, 0644); err != nil {
		return err
	}
	if err := atomicfile.Write(sysctlFile, output, 0644); err != nil {
		return fmt.Errorf("configure IPv4 forwarding: %w", err)
	}
	if err := commandRunner.Run("sysctl", "-p", sysctlFile); err != nil {
		return fmt.Errorf("enable IPv4 forwarding: %w", err)
	}
	return nil
}
