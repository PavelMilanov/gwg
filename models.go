package main

import (
	"encoding/json"
	"fmt"
	"os"
)

/*
Модель для конфигурационного файла (conf,json) сервера.
*/
type WgServerConfig struct {
	ServerPrivateKey string
	ServerPublicKey  string
	LocalAddress     string
	PublicAddress    string
	ListenPort       int
	Eth              string
	Alias            string
	Users            []UserConfig // for peer
}

/*
Генерирует вспомогательный конфигурационый файл (json) сервера для работы gwg.
*/
func (config *WgServerConfig) createServerConfigFile() {
	file, _ := json.MarshalIndent(config, "", " ")
	filename := fmt.Sprintf("%s/%s.json", WG_MANAGER_DIR, config.Alias)
	_ = os.WriteFile(filename, file, 0660)
}

/*
Модель для конфигурационных файлов (conf,json) клиентов.
*/
type UserConfig struct {
	ClientPrivateKey   string
	ClientPublicKey    string
	ClientLocalAddress string
	ServerPublicKey    string
	ServerIp           string
	ServerPort         int
	Name               string
}

/*
Генерирует вспомогательный конфигурационый файл (json) клиента для работы gwg.
*/
func (config *UserConfig) addConfigUser(fileName string) {
	file, _ := json.MarshalIndent(config, "", " ")
	filename := fmt.Sprintf("%s/%s.json", USERS_CONFIG_DIR, fileName)
	_ = os.WriteFile(filename, file, 0660)
}
