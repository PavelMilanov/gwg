package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type WgServerConfig struct {
	ServerPrivateKey string
	ServerPublicKey  string
	LocalAddress     string
	PublicAddress    string
	ListenPort       int
	Eth              string
	Alias            string
}

func (config *WgServerConfig) createServerConfigFile() {
	file, _ := json.MarshalIndent(config, "", " ")
	filename := fmt.Sprintf("%s/%s.json", WG_MANAGER_DIR, config.Alias)
	_ = os.WriteFile(filename, file, 0644)
}

type UserConfig struct {
	ClientPrivateKey   string
	ClientPublicKey    string
	ClientLocalAddress string
	ServerPublicKey    string
	ServerIp           string
	ServerPort         int
}

func (config *UserConfig) addConfigUser(fileName string) {
	file, _ := json.MarshalIndent(config, "", " ")
	filename := fmt.Sprintf("%s/%s.json", USERS_CONFIG_DIR, fileName)
	_ = os.WriteFile(filename, file, 0644)
}
