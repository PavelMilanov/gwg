package server

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/PavelMilanov/go-wg-manager/paths"
)

/*
Генерация конфигурационного файла (conf) сервера по шаблону.
*/
func writeServerConfig(config WgServerConfig, filename string) {
	serverFile := fmt.Sprintf("%s/%s.conf", paths.SERVER_DIR, filename)
	templ, err := template.New("server").Parse(SERVER_TEMPLATE)
	file, err := os.OpenFile(serverFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0660)
	err = templ.Execute(file, config)
	if err != nil {
		fmt.Println("Error writing server config")
		fmt.Println(err)
		os.Remove(serverFile)
	}
	defer file.Close()
}

/*
Генерация конфигурационного файла (conf) клиента по шаблону.
*/
func writeClientConfig(config UserConfig, filename string) {
	clientFile := fmt.Sprintf("%s/%s.conf", paths.USERS_DIR, filename)
	templ, err := template.New("client").Parse(CLIENT_TEMPLATE)
	file, err := os.OpenFile(clientFile, os.O_CREATE|os.O_WRONLY, 0660)
	err = templ.Execute(file, config)
	if err != nil {
		fmt.Println("Error writing client config")
		fmt.Println(err)
		os.Remove(clientFile)
	}
	defer file.Close()
}

/*
Чтение конфигурациионного файла сервера.
*/
func ReadServerConfigFile() WgServerConfig {
	files, _ := os.ReadDir(paths.WG_MANAGER_DIR)
	config := WgServerConfig{}
	for _, file := range files {
		content, err := os.ReadFile(paths.WG_MANAGER_DIR + "/" + file.Name())
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(content, &config)
	}
	return config
}

/*
Чтение конфигурациионных файлов клиентов.
*/
func ReadClientConfigFiles() []UserConfig {
	files, _ := os.ReadDir(paths.USERS_CONFIG_DIR)
	config := UserConfig{}
	var configs []UserConfig
	for _, file := range files {
		content, err := os.ReadFile(paths.USERS_CONFIG_DIR + "/" + file.Name())
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(content, &config)
		configs = append(configs, config)
	}
	return configs
}

/*
Добавление пользователя.
*/
func AddUSer(alias string) {
	clientPrivKey, clientPubKey := generateKeys()
	clientip := setClientIp()
	server := ReadServerConfigFile()
	config := UserConfig{
		ClientPrivateKey:   clientPrivKey,
		ClientPublicKey:    clientPubKey,
		ClientLocalAddress: clientip,
		ServerPublicKey:    server.ServerPublicKey,
		ServerIp:           server.PublicAddress,
		ServerPort:         server.ListenPort,
		DnsResolv:          server.DnsResolv,
		Name:               alias,
		Status:             "active",
	}
	config.addConfigUser(alias)
	writeClientConfig(config, alias)
	users := ReadClientConfigFiles()
	server.Users = users
	writeServerConfig(server, server.Alias)
	commandServer("restart")
	fmt.Println("User added")
}

/*
Блокировка/разблокировка пользователя.
*/
func ChangeStatusUser(alias string, state string) {
	server := ReadServerConfigFile()
	jsonfile := fmt.Sprintf("%s/%s.json", paths.USERS_CONFIG_DIR, alias)
	config := UserConfig{}
	content, err := os.ReadFile(jsonfile)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(content, &config)
	switch state {
	case "block":
		config.Status = ""
	case "unblock":
		config.Status = "active"
	}
	os.Remove(jsonfile)
	config.addConfigUser(alias)
	writeClientConfig(config, alias)
	users := ReadClientConfigFiles()
	server.Users = users
	writeServerConfig(server, server.Alias)
	commandServer("restart")
	text := fmt.Sprintf("User state changed to %s", state)
	fmt.Println(text)
}

/*
Удаление пользователя.
*/
func RemoveUser(alias string) {
	server := ReadServerConfigFile()
	configfile := fmt.Sprintf("%s/%s.conf", paths.USERS_DIR, alias)
	jsonfile := fmt.Sprintf("%s/%s.json", paths.USERS_CONFIG_DIR, alias)
	configs := []string{configfile, jsonfile}
	for _, file := range configs {
		err := os.Remove(file)
		if err != nil {
			fmt.Println(err)
		}
	}
	users := ReadClientConfigFiles()
	server.Users = users
	writeServerConfig(server, server.Alias)
	commandServer("restart")
	fmt.Println("User deleted")
}

/*
Установка Wireguard сервера.
*/
func InstallServer(alias string, network string, port int) {
	serverFile := fmt.Sprintf("%s/%s.conf", paths.SERVER_DIR, alias)
	os.Create(serverFile)
	os.Mkdir(paths.WG_MANAGER_DIR, 0660)
	privKey, pubKey := generateKeys()
	configureServer(privKey, pubKey, alias, network, port)
	commandServer("enable")
	commandServer("start")
	fmt.Println("Server started")
}

/*
Создание шаблона конфигурационного файла сервера.
*/
func configureServer(priv string, pub string, alias string, addr string, port int) {
	var (
		intf        string
		public_addr string
	)
	public_addr, intf = setServerParams()
	isValid, _ := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}/[0-9]{1,2}`, addr)
	if !isValid {
		fmt.Println("Enter valid value. Example: 10.0.0.1/24")
		os.Exit(1)
	}
	config := WgServerConfig{
		ServerPrivateKey: priv,
		ServerPublicKey:  pub,
		LocalAddress:     addr,
		PublicAddress:    public_addr,
		ListenPort:       port,
		Eth:              intf,
		DnsResolv:        "8.8.8.8",
		Alias:            alias,
	}
	writeServerConfig(config, alias)
	config.createServerConfigFile()
}

/*
Парсинг и преобразование данных из wg show dump
*/
func ReadWgDump() {
	// sudo wg show wg0 dump | wc -l = x - узнаем сколько строк в фале
	// sed -n '2,x' - выводим все, кроме 1ой
	// sed -n 'xp' - выводим определенную строку
	// command := fmt.Sprintf("wc -l dump.log")
	command := fmt.Sprintf("sudo wg show wg0 dump | wc -l")
	out, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		fmt.Println(err)
	}
	formatOut := strings.TrimRight(string(out), "\n")            // убираем \n вконце
	count, err := strconv.Atoi(strings.Split(formatOut, " ")[0]) // [8 dump.log]
	if err != nil {
		fmt.Println(err)
	}
	pool := []WireguardDump{}
	configs := ReadClientConfigFiles()
	for i := 2; i < int(count)+1; i++ {
		// command := fmt.Sprintf("sed -n '%dp' dump.log", i)
		command := fmt.Sprintf("sudo wg show wg0 dump | sed -n '%dp'", i)
		out, err := exec.Command("bash", "-c", command).Output()
		if err != nil {
			fmt.Println(err)
		}
		data := strings.Split(string(out), "\t") // Os9rBvPsb824pzh95oSyoXnGPD6jK2YKr7NK4OBoRXU=    (none)  176.59.57.104:61476     10.0.0.5/32     1695899229      816     3776    off
		var user string
		for _, config := range configs {
			if config.ClientPublicKey == data[0] { // сравниваем публичные ключи, чтобы узнать имя
				user = config.Name
				break
			}
		}
		ip := data[3]
		rateRx, err := strconv.Atoi(data[5]) // 816
		rateTx, err := strconv.Atoi(data[6]) // 3776
		pool = append(pool, WireguardDump{
			user:   user,
			ip:     ip,
			rateRx: rateRx,
			rateTx: rateTx})
	}
	for idx, line := range pool {
		text := fmt.Sprintf("%d) User: %s, Ip: %s , Resieve: %d, Sent: %d", idx+1, line.user, line.ip, line.rateRx, line.rateTx)
		fmt.Println(text)
	}
}

/*
Настраивает систему перед установкой gwg.
*/
func ConfigureSystem() {
	initSystem()
	installFile := "./setup.sh"
	err := os.WriteFile(installFile, []byte(GWG_UTILS), 0751)
	if err != nil {
		fmt.Println(err)
	}
	command := fmt.Sprintf("%s install", installFile)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	defer os.Remove(installFile)
}
