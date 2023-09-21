package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type WgServerConfig struct {
	/*
		Модель для конфигурационного файла (conf,json) сервера.
	*/
	ServerPrivateKey string
	ServerPublicKey  string
	LocalAddress     string
	PublicAddress    string
	ListenPort       int
	Eth              string
	Alias            string
	Users            []string // for peer
}

func (config *WgServerConfig) createServerConfigFile() {
	/*
		Генерирует вспомогательный конфигурационый файл сервера для работы gwg.
	*/
	file, _ := json.MarshalIndent(config, "", " ")
	filename := fmt.Sprintf("%s/%s.json", WG_MANAGER_DIR, config.Alias)
	_ = os.WriteFile(filename, file, 0644)
}

type UserConfig struct {
	/*
		Модель для конфигурационных файлов (conf,json) клиентов.
	*/
	ClientPrivateKey   string
	ClientPublicKey    string
	ClientLocalAddress string
	ServerPublicKey    string
	ServerIp           string
	ServerPort         int
}

func (config *UserConfig) addConfigUser(fileName string) {
	/*
		Генерирует вспомогательный конфигурационый файл клиента для работы gwg.
	*/
	file, _ := json.MarshalIndent(config, "", " ")
	filename := fmt.Sprintf("%s/%s.json", USERS_CONFIG_DIR, fileName)
	_ = os.WriteFile(filename, file, 0644)
}
