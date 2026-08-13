package server

import (
	"encoding/json"
	"fmt"

	"github.com/PavelMilanov/go-wg-manager/internal/atomicfile"
	"github.com/PavelMilanov/go-wg-manager/paths"
)

var (
	serverDir     = paths.SERVER_DIR
	managerDir    = paths.WG_MANAGER_DIR
	userConfigDir = paths.USERS_CONFIG_DIR
	usersDir      = paths.USERS_DIR
	tcConfigDir   = paths.TC_DIR
	sysctlFile    = "/etc/sysctl.d/90-gwg.conf"
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
	DnsResolv          string
	Name               string
	Status             string
}

/*
Генерирует вспомогательный конфигурационый файл (json) клиента для работы gwg.
*/
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

/*
Модель для парсинга данных из wg show dump.
*/
type WireguardDump struct {
	user   string
	ip     string
	rateRx uint64
	rateTx uint64
}
