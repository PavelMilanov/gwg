package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func generateKeys() (string, string) {
	/*
		Генерация приватного и публичного ключей сервера и сохранение в файлы.
	*/
	dir := os.TempDir()
	os.Chdir(dir)
	fmt.Println("Generate keys...")
	cmd := exec.Command("bash", "-c", "wg genkey | tee privatekey | wg pubkey | tee publickey")
	cmd.Stderr = os.Stderr
	cmd.Run()
	privatekey, _ := os.ReadFile("privatekey")
	publickey, _ := os.ReadFile("publickey")
	defer os.RemoveAll(dir)
	return string(privatekey), string(publickey)
}

func readConfigFile(path string) {
	switch path {
	case WG_MANAGER_DIR:
		configs, _ := os.ReadDir(WG_MANAGER_DIR)
		for _, config := range configs {
			model := WgServerConfigFile{}
			content, err := os.ReadFile(config.Name())
			if err != nil {
				panic(err)
			}
			json.Unmarshal(content, &model)
			fmt.Println(config)
		}
		config := WgServerConfigFile{}
		filename := fmt.Sprintf("%s.json", config.Alias)
		content, err := os.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(content, &config)
		fmt.Println(config)

	}
}
