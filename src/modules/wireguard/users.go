package wireguard

import (
	"errors"
	"fmt"
	"os"
)

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

func ListUsers() error {
	users, err := ReadClientConfigFiles()
	if err != nil {
		return err
	}
	for _, user := range users {
		status := user.Status
		if status == "" {
			status = "blocked"
		}
		fmt.Printf("%s\t%s\t%s\n", user.Name, user.ClientLocalAddress, status)
	}
	return nil
}
