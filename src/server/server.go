package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/PavelMilanov/go-wg-manager/internal/atomicfile"
)

func writeServerConfig(config WgServerConfig, filename string) error {
	if err := validateInterfaceName(filename); err != nil {
		return err
	}
	if err := validateServerConfig(config); err != nil {
		return err
	}
	return writeTemplate(configPath(serverDir, filename, ".conf"), SERVER_TEMPLATE, config, 0600)
}

func writeClientConfig(config UserConfig, filename string) error {
	if err := validateUserName(filename); err != nil {
		return err
	}
	if err := validateUserConfig(config); err != nil {
		return err
	}
	return writeTemplate(configPath(usersDir, filename, ".conf"), CLIENT_TEMPLATE, config, 0600)
}

func writeTemplate(path, source string, value any, perm os.FileMode) error {
	templ, err := template.New(filepath.Base(path)).Parse(source)
	if err != nil {
		return fmt.Errorf("parse template for %s: %w", path, err)
	}
	var output bytes.Buffer
	if err := templ.Execute(&output, value); err != nil {
		return fmt.Errorf("render template for %s: %w", path, err)
	}
	return atomicfile.Write(path, output.Bytes(), perm)
}

func ReadServerConfigFile() (WgServerConfig, error) {
	entries, err := os.ReadDir(managerDir)
	if err != nil {
		return WgServerConfig{}, fmt.Errorf("read server configuration directory: %w", err)
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			files = append(files, entry.Name())
		}
	}
	if len(files) == 0 {
		return WgServerConfig{}, fmt.Errorf("server is not configured")
	}
	if len(files) > 1 {
		return WgServerConfig{}, fmt.Errorf("multiple server configurations found in %s", managerDir)
	}

	var config WgServerConfig
	if err := readJSON(filepath.Join(managerDir, files[0]), &config); err != nil {
		return WgServerConfig{}, err
	}
	if err := validateServerConfig(config); err != nil {
		return WgServerConfig{}, fmt.Errorf("invalid stored server configuration: %w", err)
	}
	return config, nil
}

func ReadClientConfigFiles() ([]UserConfig, error) {
	entries, err := os.ReadDir(userConfigDir)
	if err != nil {
		return nil, fmt.Errorf("read user configuration directory: %w", err)
	}
	configs := make([]UserConfig, 0, len(entries))
	addresses := make(map[string]string, len(entries))
	publicKeys := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		var config UserConfig
		if err := readJSON(filepath.Join(userConfigDir, entry.Name()), &config); err != nil {
			return nil, err
		}
		if err := validateUserName(config.Name); err != nil {
			return nil, fmt.Errorf("invalid stored user configuration %s: %w", entry.Name(), err)
		}
		if strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) != config.Name {
			return nil, fmt.Errorf("user configuration filename %s does not match user %q", entry.Name(), config.Name)
		}
		if err := validateUserConfig(config); err != nil {
			return nil, fmt.Errorf("invalid stored user configuration %s: %w", entry.Name(), err)
		}
		if owner, exists := addresses[config.ClientLocalAddress]; exists {
			return nil, fmt.Errorf("users %q and %q share address %s", owner, config.Name, config.ClientLocalAddress)
		}
		if owner, exists := publicKeys[config.ClientPublicKey]; exists {
			return nil, fmt.Errorf("users %q and %q share a public key", owner, config.Name)
		}
		addresses[config.ClientLocalAddress] = config.Name
		publicKeys[config.ClientPublicKey] = config.Name
		configs = append(configs, config)
	}
	return configs, nil
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func AddUser(alias string) error {
	if err := validateUserName(alias); err != nil {
		return err
	}
	server, err := ReadServerConfigFile()
	if err != nil {
		return err
	}
	users, err := ReadClientConfigFiles()
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == alias {
			return fmt.Errorf("user %q already exists", alias)
		}
	}
	clientIP, err := setClientIP(server, users)
	if err != nil {
		return err
	}
	clientPrivateKey, clientPublicKey, err := generateKeys()
	if err != nil {
		return err
	}
	config := UserConfig{
		ClientPrivateKey:   clientPrivateKey,
		ClientPublicKey:    clientPublicKey,
		ClientLocalAddress: clientIP,
		ServerPublicKey:    server.ServerPublicKey,
		ServerIp:           server.PublicAddress,
		ServerPort:         server.ListenPort,
		DnsResolv:          server.DnsResolv,
		Name:               alias,
		Status:             "active",
	}
	if err := config.addConfigUser(alias); err != nil {
		return err
	}
	if err := writeClientConfig(config, alias); err != nil {
		_ = os.Remove(configPath(userConfigDir, alias, ".json"))
		return err
	}
	if err := syncServerUsers(&server); err != nil {
		return err
	}
	if err := commandServer("restart", server.Alias); err != nil {
		return err
	}
	fmt.Println("User added")
	return nil
}

func ChangeStatusUser(alias, state string) error {
	if err := validateUserName(alias); err != nil {
		return err
	}
	if state != "block" && state != "unblock" {
		return fmt.Errorf("invalid user state %q", state)
	}
	server, err := ReadServerConfigFile()
	if err != nil {
		return err
	}
	var config UserConfig
	jsonFile := configPath(userConfigDir, alias, ".json")
	if err := readJSON(jsonFile, &config); err != nil {
		return err
	}
	if config.Name != alias {
		return fmt.Errorf("user configuration %s belongs to %q", jsonFile, config.Name)
	}
	if state == "block" {
		config.Status = ""
	} else {
		config.Status = "active"
	}
	if err := config.addConfigUser(alias); err != nil {
		return err
	}
	if err := syncServerUsers(&server); err != nil {
		return err
	}
	if err := commandServer("restart", server.Alias); err != nil {
		return err
	}
	fmt.Printf("User state changed to %s\n", state)
	return nil
}

func RemoveUser(alias string) error {
	if err := validateUserName(alias); err != nil {
		return err
	}
	server, err := ReadServerConfigFile()
	if err != nil {
		return err
	}
	jsonFile := configPath(userConfigDir, alias, ".json")
	if _, err := os.Stat(jsonFile); err != nil {
		return fmt.Errorf("user %q not found: %w", alias, err)
	}
	for _, path := range []string{configPath(usersDir, alias, ".conf"), jsonFile} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", path, err)
		}
	}
	if err := syncServerUsers(&server); err != nil {
		return err
	}
	if err := commandServer("restart", server.Alias); err != nil {
		return err
	}
	fmt.Println("User deleted")
	return nil
}

func syncServerUsers(server *WgServerConfig) error {
	users, err := ReadClientConfigFiles()
	if err != nil {
		return err
	}
	server.Users = users
	if err := writeServerConfig(*server, server.Alias); err != nil {
		return err
	}
	return server.createServerConfigFile()
}

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
	for _, dir := range []string{serverDir, managerDir, userConfigDir, usersDir} {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
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

func ReadWgDump() error {
	server, err := ReadServerConfigFile()
	if err != nil {
		return err
	}
	users, err := ReadClientConfigFiles()
	if err != nil {
		return err
	}
	out, err := commandRunner.Output("sudo", "wg", "show", server.Alias, "dump")
	if err != nil {
		return err
	}
	pool, err := parseDump(out, users)
	if err != nil {
		return err
	}
	for index, line := range pool {
		fmt.Printf("%d) User: %s, IP: %s, Received: %d, Sent: %d\n", index+1, line.user, line.ip, line.rateRx, line.rateTx)
	}
	return nil
}

func ConfigureSystem() error {
	if err := initSystem(); err != nil {
		return err
	}
	if os.Geteuid() != 0 {
		return fmt.Errorf("initialization requires root privileges; run sudo gwg init")
	}
	if err := prepareSystem(); err != nil {
		return err
	}
	return InstallServer("wg0", "10.0.0.1/24", 51830)
}

func prepareSystem() error {
	for _, dir := range []string{serverDir, managerDir, userConfigDir, usersDir, tcConfigDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
		if err := os.Chmod(dir, 0700); err != nil {
			return fmt.Errorf("set permissions on directory %s: %w", dir, err)
		}
	}
	if err := atomicfile.Write(sysctlFile, []byte("net.ipv4.ip_forward=1\n"), 0644); err != nil {
		return fmt.Errorf("configure IPv4 forwarding: %w", err)
	}
	if err := commandRunner.Run("sysctl", "-p", sysctlFile); err != nil {
		return fmt.Errorf("enable IPv4 forwarding: %w", err)
	}
	return nil
}
