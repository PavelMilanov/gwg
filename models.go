package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type WgServerConfigFile struct {
	PublicKey     string
	PrivateKey    string
	LocalAddress  string
	PublicAddress string
	ListenPort    int
	Alias         string
}

func (config *WgServerConfigFile) createServerConfigFile() {
	file, _ := json.MarshalIndent(config, "", " ")
	os.Chdir(WG_MANAGER_DIR)
	filename := fmt.Sprintf("%s.json", config.Alias)
	_ = os.WriteFile(filename, file, 0644)
}

type UserConfig struct {
	ClientPrivateKey   string
	ClientLocalAddress string
	ServerPublicKey    string
	ServerIp           string
	ServerPort         int
}

func (config *UserConfig) addConfigUser(fileName string) {
	file, _ := json.MarshalIndent(config, "", " ")
	fmt.Println("Enter client description:")
	os.Chdir(USERS_CONFIG_DIR)
	filename := fmt.Sprintf("%s.json", fileName)
	_ = os.WriteFile(filename, file, 0644)
}
