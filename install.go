package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"text/template"
)

type WgServerConfig struct {
	ServerPrivateKey string
	LocalAddress     string
	ListenPort       int
	Eth              string
}

type WgServerConfigFile struct {
	PublicKey     string
	PrivateKey    string
	LocalAddress  string
	PublicAddress string
	ListenPort    int
	Alias         string
}

func (config *WgServerConfigFile) createConfigFile() {
	file, _ := json.MarshalIndent(config, "", " ")
	os.Chdir(WG_MANAGER_DIR)
	filename := fmt.Sprintf("%s.json", config.Alias)
	_ = os.WriteFile(filename, file, 0644)
}

func installServer() {
	/*
		Основаня логика установки WG Server.
	*/
	updatePackage()
	installWgServer()
	os.Mkdir(WG_MANAGER_DIR, 0666)
	generateKeys()
	configureServer()
}
func updatePackage() {
	/*
		Обновление пакетов deb.
	*/
	fmt.Println("Updating packages...")
	cmd := exec.Command("apt", "update", "-y")
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func installWgServer() {
	/*
		Установка пакета wireguard.
	*/
	fmt.Println("Installing WireGuard Server...")
	cmd := exec.Command("apt", "install", "-y", "wireguard")
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func generateKeys() {
	/*
		Генерация приватного и публичного ключей сервера и сохранение в файлы.
	*/
	os.Chdir(WG_MANAGER_DIR)
	fmt.Println("Generate keys...")
	cmd := exec.Command("bash", "-c", "wg genkey | tee privatekey | wg pubkey | tee publickey")
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func configureServer() {
	/*
		Создание шаблона конфигурационного файла сервера.
	*/
	var (
		private_addr string
		port         int
		intf         string
		alias        string
	)
	fmt.Println("Enter private network: 10.0.0.1/24")
	private_addr_value, _ := fmt.Scanf("%s\n", &private_addr)
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
	port_value, _ := fmt.Scanf("%d\n", &port)
	if port_value == 0 {
		port = 51830
	}
	fmt.Println("Enter NAT-interface:")
	intf_value, _ := fmt.Scanf("%s\n", &intf)
	if intf_value == 0 {
		fmt.Println("Enter NAT-interface")
		os.Exit(1)
	}
	fmt.Println("Enter alias: 'wg0'")
	alias_value, _ := fmt.Scanf("%s\n", &alias)
	if alias_value == 0 {
		alias = "wg0"
	}
	privatekeypath := fmt.Sprintf("%s/privatekey", WG_MANAGER_DIR)
	publickeypath := fmt.Sprintf("%s/publickey", WG_MANAGER_DIR)
	privatekey, _ := os.ReadFile(privatekeypath)
	publickey, _ := os.ReadFile(publickeypath)
	config := WgServerConfig{
		ServerPrivateKey: string(privatekey),
		LocalAddress:     private_addr,
		ListenPort:       port,
		Eth:              intf,
	}
	serverFile := fmt.Sprintf("%s/%s.conf", SERVER_DIR, alias)
	templ, err := template.ParseFiles("wg_template.conf")
	file, err := os.OpenFile(serverFile, os.O_CREATE|os.O_WRONLY, 0666)
	err = templ.Execute(file, config)
	if err != nil {
		panic(err)
	}
	configFile := WgServerConfigFile{
		PublicKey:    string(publickey),
		PrivateKey:   string(privatekey),
		LocalAddress: private_addr,
		ListenPort:   port,
		Alias:        alias,
	}
	configFile.createConfigFile()
	defer file.Close()
}
