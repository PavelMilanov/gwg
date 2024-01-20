package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PavelMilanov/go-wg-manager/paths"
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
	DnsResolv        string
	Users            []UserConfig // for peer
}

/*
Генерирует вспомогательный конфигурационый файл (json) сервера для работы gwg.
*/
func (config *WgServerConfig) createServerConfigFile() {
	file, _ := json.MarshalIndent(config, "", "\t")
	filename := fmt.Sprintf("%s/%s.json", paths.WG_MANAGER_DIR, config.Alias)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
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
	Status             string
}

/*
Генерирует вспомогательный конфигурационый файл (json) клиента для работы gwg.
*/
func (config *UserConfig) addConfigUser(fileName string) {
	file, _ := json.MarshalIndent(config, "", "\t")
	filename := fmt.Sprintf("%s/%s.json", paths.USERS_CONFIG_DIR, fileName)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
}

/*
Модель для парсинга данных из wg show dump.
*/
type WireguardDump struct {
	user   string
	ip     string
	rateRx int
	rateTx int
}
