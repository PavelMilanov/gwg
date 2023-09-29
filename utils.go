package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"text/template"
)

const (
	SERVER_TEMPLATE = WG_MANAGER_DIR + "/wg_template.conf"
	CLIENT_TEMPLATE = WG_MANAGER_DIR + "/client_template.conf"
)

/*
Генерация конфигурационного файла (conf) сервера по шаблону.
*/
func writeServerConfig(config WgServerConfig, filename string) {
	serverFile := fmt.Sprintf("%s/%s.conf", SERVER_DIR, filename)
	templ, err := template.ParseFiles(SERVER_TEMPLATE)
	file, err := os.OpenFile(serverFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0660)
	err = templ.Execute(file, config)
	if err != nil {
		fmt.Println(err)
		os.Remove(serverFile)
		os.Exit(1)
	}
	fmt.Println("Done writing server config")
	defer file.Close()
}

/*
Генерация конфигурационного файла (conf) клиента по шаблону.
*/
func writeClientConfig(config UserConfig, filename string) {
	clientFile := fmt.Sprintf("%s/%s.conf", USERS_DIR, filename)
	clientTemplate, err := template.ParseFiles(CLIENT_TEMPLATE)
	file, err := os.OpenFile(clientFile, os.O_CREATE|os.O_WRONLY, 0660)
	err = clientTemplate.Execute(file, config)
	if err != nil {
		fmt.Println(err)
		os.Remove(clientFile)
		os.Exit(1)
	}
	fmt.Println("Done writing client config")
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
Добавление пользователя.
*/
func addUSer(alias string) {
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
	restartServer()
}

/*
Удаление пользователя.
*/
func removeUser(alias string) {
	server := readServerConfigFile()
	configfile := fmt.Sprintf("%s/%s.conf", USERS_DIR, alias)
	jsonfile := fmt.Sprintf("%s/%s.json", USERS_CONFIG_DIR, alias)
	configs := []string{configfile, jsonfile}
	for _, file := range configs {
		err := os.Remove(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	users := readClientConfigFiles()
	server.Users = users
	writeServerConfig(server, server.Alias)
	restartServer()
}

/*
Установка Wireguard сервера.
*/
func installServer(alias string) {
	serverFile := fmt.Sprintf("%s/%s.conf", SERVER_DIR, alias)
	os.Create(serverFile)
	os.Mkdir(WG_MANAGER_DIR, 0660)
	privKey, pubKey := generateKeys()
	configureServer(privKey, pubKey, alias)
}

/*
Создание шаблона конфигурационного файла сервера.
*/
func configureServer(priv string, pub string, alias string) {
	var (
		private_addr string
		port         int
		intf         string
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
