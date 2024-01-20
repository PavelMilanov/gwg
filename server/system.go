package server

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

/*
Проверка операционной системы на совместимость.
*/
func initSystem() {
	_, err := exec.Command("bash", "-c", "cat /etc/os-release").Output()
	if err != nil {
		fmt.Println("Operating system does not support")
		os.Exit(1)
	}
}

/*
Динамическое назначение приватных ip-адресов клиентам.
*/
func setClientIp() string {
	configs := ReadClientConfigFiles()
	var pattern = 2 // выдавать адреса с 2
	var ipv4 string
	if len(configs) == 0 {
		server := ReadServerConfigFile()
		serverIPv4 := server.LocalAddress[:len(server.LocalAddress)-3] // 10.0.0.1 -=> 10.0.0.1/32
		clientIPv4 := strings.Split(serverIPv4, ".")
		IPv4byte1, _ := strconv.Atoi(clientIPv4[0])
		IPv4byte2, _ := strconv.Atoi(clientIPv4[1])
		IPv4byte3, _ := strconv.Atoi(clientIPv4[2])
		IPv4byte4, _ := strconv.Atoi(clientIPv4[3])
		ipv4 = fmt.Sprintf("%d.%d.%d.%d/32", IPv4byte1, IPv4byte2, IPv4byte3, IPv4byte4+1) // 10.0.0.1 => 10.0.0.2/32
	} else {
		for index, config := range configs {
			data := config.ClientLocalAddress[:len(config.ClientLocalAddress)-3] // 10.0.0.5 => 10.0.0.5/32
			clientIPv4 := strings.Split(data, ".")
			IPv4byte1, _ := strconv.Atoi(clientIPv4[0])
			IPv4byte2, _ := strconv.Atoi(clientIPv4[1])
			IPv4byte3, _ := strconv.Atoi(clientIPv4[2])
			IPv4byte4, _ := strconv.Atoi(clientIPv4[3])
			if pattern < IPv4byte4 {
				ipv4 = fmt.Sprintf("%d.%d.%d.%d/32", IPv4byte1, IPv4byte2, IPv4byte3, pattern)
				break
			}
			if index+1 == len(configs) {
				ipv4 = fmt.Sprintf("%d.%d.%d.%d/32", IPv4byte1, IPv4byte2, IPv4byte3, pattern+1)
			}
			pattern++
		}
	}
	return ipv4
}

/*
Автопоиск интерфейса и ip для конфигурации сервера.
*/
func setServerParams() (string, string) {
	out, err := exec.Command("bash", "-c", "ip r").Output()
	if err != nil {
		fmt.Println(err)
	}
	var serverIp, serverIntf string
	defaultRoute := strings.Split(string(out), " ")[:5] // первая строка "default via 192.168.11.1 dev vlan601 proto static metric 404 ..."
	ip := defaultRoute[2]
	gate4 := net.ParseIP(ip)
	serverIntf = defaultRoute[4]
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
	}
	for _, interf := range interfaces {
		// Список адресов для каждого сетевого интерфейса
		addrs, err := interf.Addrs()
		if err != nil {
			fmt.Println(err)
		}
		for _, addr := range addrs {
			data := addr.String()
			ip, ipnet, _ := net.ParseCIDR(data)
			if ipnet.Contains(gate4) {
				serverIp = ip.String()
			}
		}
	}
	return serverIp, serverIntf
}

/*
Генерация приватного и публичного ключей.
*/
func generateKeys() (string, string) {
	dir := os.TempDir()
	os.Chdir(dir)
	cmd := exec.Command("bash", "-c", "wg genkey | tee privatekey | wg pubkey | tee publickey")
	cmd.Stderr = os.Stderr
	cmd.Run()
	privatekeyToFile, _ := os.ReadFile("privatekey")
	publickeyToFile, _ := os.ReadFile("publickey")
	privatekey := strings.TrimRight(string(privatekeyToFile), "\n")
	publickey := strings.TrimRight(string(publickeyToFile), "\n")
	defer os.RemoveAll(dir)
	return privatekey, publickey
}

/*
Просмотр статистики wg.
*/
func ShowPeers() {
	command := fmt.Sprintf("sudo wg show")
	out, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))
}

/*
Управление службой wg-quick.
*/
func commandServer(cmd string) {
	server := ReadServerConfigFile()
	command := fmt.Sprintf("sudo systemctl %s wg-quick@%s.service", cmd, server.Alias)
	_, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		fmt.Println(err)
	}
}
