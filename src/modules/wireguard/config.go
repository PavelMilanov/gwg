package wireguard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PavelMilanov/go-wg-manager/internal/atomicfile"
	"github.com/PavelMilanov/go-wg-manager/internal/layout"
)

const defaultServerDir = "/etc/wireguard"

var (
	serverDir     = defaultServerDir
	managerDir    = layout.WireGuardDir
	userConfigDir = layout.WireGuardDir + "/clients"
	usersDir      = layout.WireGuardDir + "/configs"
	sysctlFile    = "/etc/sysctl.d/90-gwu.conf"
)

func writeServerConfig(config WgServerConfig, filename string) error {
	if err := validateInterfaceName(filename); err != nil {
		return err
	}
	if err := validateServerConfig(config); err != nil {
		return err
	}
	output, err := serverTemplate.render(config)
	if err != nil {
		return err
	}
	if err := writeGenerated(filename+".conf", output, 0600); err != nil {
		return err
	}
	return atomicfile.Write(configPath(serverDir, filename, ".conf"), output, 0600)
}

func writeGenerated(name string, data []byte, perm os.FileMode) error {
	dir := filepath.Join(managerDir, "generated")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return atomicfile.Write(filepath.Join(dir, name), data, perm)
}

func writeClientConfig(config UserConfig, filename string) error {
	if err := validateUserName(filename); err != nil {
		return err
	}
	if err := validateUserConfig(config); err != nil {
		return err
	}
	return writeTemplate(configPath(usersDir, filename, ".conf"), clientTemplate, config, 0600)
}

func writeTemplate(path string, tmpl configTemplate, value any, perm os.FileMode) error {
	output, err := tmpl.render(value)
	if err != nil {
		return err
	}
	return atomicfile.Write(path, output, perm)
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

func (config *WgServerConfig) createServerConfigFile() error {
	if err := validateServerConfig(*config); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal server configuration: %w", err)
	}
	filename := configPath(managerDir, config.Alias, ".json")
	return atomicfile.Write(filename, data, 0600)
}

func (config *UserConfig) addConfigUser(fileName string) error {
	if err := validateUserConfig(*config); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal user configuration: %w", err)
	}
	filename := configPath(userConfigDir, fileName, ".json")
	return atomicfile.Write(filename, data, 0600)
}
