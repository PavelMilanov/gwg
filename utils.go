package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"text/template"
)

/*
Генерация конфигурационного файла (conf) сервера по шаблону.
*/
func writeServerConfig(config WgServerConfig, filename string) {
	serverFile := fmt.Sprintf("%s/%s.conf", SERVER_DIR, filename)
	templ, err := template.ParseFiles("./wg_template.conf")
	file, err := os.OpenFile(serverFile, os.O_CREATE|os.O_WRONLY, 0666)
	err = templ.Execute(file, config)
	if err != nil {
		panic(err)
	}
	defer file.Close()
}

/*
Генерация конфигурационного файла (conf) клиента по шаблону.
*/
func writeClientConfig(config UserConfig, filename string) {
	clientFile := fmt.Sprintf("%s/%s.conf", USERS_DIR, filename)
	clientTemplate, err := template.ParseFiles("./client_template.conf")
	file, err := os.OpenFile(clientFile, os.O_CREATE|os.O_WRONLY, 0666)
	err = clientTemplate.Execute(file, config)
	if err != nil {
		panic(err)
	}
	defer file.Close()
}

/*
Чтение конфигурациионного файла сервера.
*/
func readServerConfigFile() WgServerConfig {
	files, _ := os.ReadDir(WG_MANAGER_DIR)
	config := WgServerConfig{}
	for _, file := range files {
		content, err := os.ReadFile(WG_MANAGER_DIR + "/" + file.Name())
		if err != nil {
			panic(err)
		}
		json.Unmarshal(content, &config)
	}
	return config
}

/*
Чтение конфигурациионных файлов клиентов.
*/
func readClientConfigFiles() []UserConfig {
	files, _ := os.ReadDir(USERS_CONFIG_DIR)
	config := UserConfig{}
	var configs []UserConfig
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

/*
Основная логика при вводе команды add.
*/
func addUSer() {
	var alias string
	fmt.Println("Enter client description:")
	alias_value, _ := fmt.Scanf("%s", &alias)
	if alias_value == 0 {
		os.Exit(1)
	}
	clientPrivKey, clientPubKey := generateKeys()
	clientip := setClientIp()
	server := readServerConfigFile()
	config := UserConfig{
		ClientPrivateKey:   clientPrivKey,
		ClientPublicKey:    clientPubKey,
		ClientLocalAddress: clientip,
		ServerPublicKey:    server.ServerPublicKey,
		ServerIp:           server.PublicAddress,
		ServerPort:         server.ListenPort,
		Name:               alias,
	}
	config.addConfigUser(alias)
	writeClientConfig(config, alias)
	users := readClientConfigFiles()
	server.Users = users
	writeServerConfig(server, server.Alias)
}

/*
Основная логика при вводе команды install.
*/
func installServer() {
	installWgServer()
	os.Mkdir(WG_MANAGER_DIR, 0666)
	privKey, pubKey := generateKeys()
	configureServer(privKey, pubKey)
}

/*
Создание шаблона конфигурационного файла сервера.
*/
func configureServer(priv string, pub string) {
	var (
		private_addr string
		port         int
		intf         string
		alias        string
		public_addr  string
	)
	public_addr, intf = setServerParams()
	fmt.Println("Enter private network: 10.0.0.1/24")
	private_addr_value, _ := fmt.Scanf("%s\r", &private_addr)
	if private_addr_value == 0 {
		private_addr = "10.0.0.1/24"
	} else {
		isValid, _ := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}/[0-9]{1,2}`, private_addr)
		if !isValid {
			fmt.Println("Enter valid value. Example: 10.0.0.1/24")
			os.Exit(1)
		}
	}
	fmt.Println("Enter listen port: 51830")
	port_value, _ := fmt.Scanf("%d\r", &port)
	if port_value == 0 {
		port = 51830
	}
	fmt.Println("Enter alias: 'wg0'")
	alias_value, _ := fmt.Scanf("%s\r", &alias)
	if alias_value == 0 {
		alias = "wg0"
	}
	config := WgServerConfig{
		ServerPrivateKey: priv,
		ServerPublicKey:  pub,
		LocalAddress:     private_addr,
		PublicAddress:    public_addr,
		ListenPort:       port,
		Eth:              intf,
		Alias:            alias,
	}
	writeServerConfig(config, alias)
	config.createServerConfigFile()
}
