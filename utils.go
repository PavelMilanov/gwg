package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func generateKeys() (string, string) {
	/*
		Генерация приватного и публичного ключей.
	*/
	dir := os.TempDir()
	os.Chdir(dir)
	fmt.Println("Generate keys...")
	cmd := exec.Command("bash", "-c", "wg genkey | tee privatekey | wg pubkey | tee publickey")
	//cmd := exec.Command("bash", "-c", "echo privatekey > privatekey && echo publickey > publickey") // testing
	cmd.Stderr = os.Stderr
	cmd.Run()
	privatekey, _ := os.ReadFile("privatekey")
	publickey, _ := os.ReadFile("publickey")
	defer os.RemoveAll(dir)
	return string(privatekey), string(publickey)
}

func readServerConfigFile() *WgServerConfig {
	files, _ := os.ReadDir(WG_MANAGER_DIR)
	config := &WgServerConfig{}
	for _, file := range files {
		content, err := os.ReadFile(WG_MANAGER_DIR + "/" + file.Name())
		if err != nil {
			panic(err)
		}
		json.Unmarshal(content, &config)
	}
	return config
}

func readClientConfigFiles() []*UserConfig {
	files, _ := os.ReadDir(USERS_CONFIG_DIR)
	config := &UserConfig{}
	var configs []*UserConfig
	for _, file := range files {
		content, err := os.ReadFile(USERS_CONFIG_DIR + "/" + file.Name())
		if err != nil {
			panic(err)
		}
		json.Unmarshal(content, &config)
		configs = append(configs, config)
	}
	return configs
}
