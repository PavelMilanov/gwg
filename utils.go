package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

/*
Обработчик ошибок.
*/
func check(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
}

/*
Генерация конфигурационного файла (conf) сервера по шаблону.
*/
func writeServerConfig(config WgServerConfig, filename string) {
	serverFile := fmt.Sprintf("%s/%s.conf", SERVER_DIR, filename)
	templ, err := template.New("server").Parse(SERVER_TEMPLATE)
	file, err := os.OpenFile(serverFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0660)
	err = templ.Execute(file, config)
	if err != nil {
		fmt.Println(err)
		os.Remove(serverFile)
		// os.Exit(1)
	}
	fmt.Println("Done writing server config")
	defer file.Close()
}

/*
Генерация конфигурационного файла (conf) клиента по шаблону.
*/
func writeClientConfig(config UserConfig, filename string) {
	clientFile := fmt.Sprintf("%s/%s.conf", USERS_DIR, filename)
	// clientTemplate, err := template.ParseFiles(CLIENT_TEMPLATE)
	templ, err := template.New("client").Parse(CLIENT_TEMPLATE)
	file, err := os.OpenFile(clientFile, os.O_CREATE|os.O_WRONLY, 0660)
	err = templ.Execute(file, config)
	if err != nil {
		fmt.Println(err)
		os.Remove(clientFile)
		// os.Exit(1)
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
		check(err)
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
		check(err)
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
		Status:             "active",
	}
	config.addConfigUser(alias)
	writeClientConfig(config, alias)
	users := readClientConfigFiles()
	server.Users = users
	writeServerConfig(server, server.Alias)
	commandServer("restart")
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
		check(err)
	}
	users := readClientConfigFiles()
	server.Users = users
	writeServerConfig(server, server.Alias)
	commandServer("restart")
}

/*
Установка Wireguard сервера.
*/
func installServer(alias string) {
	createProjectDirs()
	serverFile := fmt.Sprintf("%s/%s.conf", SERVER_DIR, alias)
	os.Create(serverFile)
	os.Mkdir(WG_MANAGER_DIR, 0660)
	privKey, pubKey := generateKeys()
	configureServer(privKey, pubKey, alias)
	commandServer("enable")
	commandServer("start")
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

/*
Парсинг и преобразование данных из wg show dump
*/
func readWgDump() {
	// sudo wg show wg0 dump | wc -l = x - узнаем сколько строк в фале
	// sed -n '2,x' - выводим все, кроме 1ой
	// sed -n 'xp' - выводим определенную строку
	// command := fmt.Sprintf("wc -l dump.log")
	command := fmt.Sprintf("sudo wg show wg0 dump | wc -l")
	out, err := exec.Command("bash", "-c", command).Output()
	check(err)
	count, err := strconv.Atoi(strings.Split(string(out), " ")[0]) // [8 dump.log]
	check(err)
	pool := []WireguardDump{}
	for i := 2; i < count+1; i++ {
		// command := fmt.Sprintf("sed -n '%dp' dump.log", i)
		command := fmt.Sprintf("sudo wg show wg0 dump | sed -n '%dp'", i)
		out, err := exec.Command("bash", "-c", command).Output()
		check(err)
		data := strings.Split(string(out), "\t") // Os9rBvPsb824pzh95oSyoXnGPD6jK2YKr7NK4OBoRXU=    (none)  176.59.57.104:61476     10.0.0.5/32     1695899229      816     3776    off
		user := data[0]                          // Os9rBvPsb824pzh95oSyoXnGPD6jK2YKr7NK4OBoRXU
		rateRx, err := strconv.Atoi(data[5])     // 816
		rateTx, err := strconv.Atoi(data[6])     // 3776
		pool = append(pool, WireguardDump{
			user:   user,
			rateRx: rateRx,
			rateTx: rateTx})
	}
	for idx, line := range pool {
		text := fmt.Sprintf("%d) User: %s, RateRx: %d, RateTx: %d", idx+1, line.user, line.rateRx, line.rateTx)
		fmt.Println(text)
	}
}
