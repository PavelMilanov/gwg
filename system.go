package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	SERVER_DIR       = "/etc/wireguard/"
	WG_MANAGER_DIR   = SERVER_DIR + ".wg_manager"
	USERS_CONFIG_DIR = SERVER_DIR + ".configs"
	USERS_DIR        = SERVER_DIR + "users"
)

/*
Проверка операционной системы на совместимость.
*/
func initSystem() {
	_, err := exec.Command("bash", "-c", "cat /etc/os-release").Output()
	check(err)
}

/*
Создает рабочие директории.
*/
func createProjectDirs() {
	err := os.Chdir(SERVER_DIR)
	fmt.Println("Creating project directories...")
	check(err)
	dirs := [3]string{WG_MANAGER_DIR, USERS_CONFIG_DIR, USERS_DIR}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0770)
		check(err)
	}
	fmt.Println("Done.")
}

/*
Динамическое назначение приватных ip-адресов клиентам.
*/
func setClientIp() string {
	configs := readClientConfigFiles()
	var pattern = 2
	var ipv4 string
	if len(configs) == 0 {
		ipv4 = "10.0.0.2/32"
	}
	for index, config := range configs {
		data := config.ClientLocalAddress[:len(config.ClientLocalAddress)-3] // 10.0.0.5 10.0.0.5/32
		clientIPv4 := strings.Split(data, ".")
		IPv4byte1, _ := strconv.ParseInt(clientIPv4[0], 10, 0)
		IPv4byte2, _ := strconv.ParseInt(clientIPv4[1], 10, 0)
		IPv4byte3, _ := strconv.ParseInt(clientIPv4[2], 10, 0)
		IPv4byte4, _ := strconv.ParseInt(clientIPv4[3], 10, 0)
		fmt.Println(pattern, IPv4byte4)
		if pattern < int(IPv4byte4) {
			ipv4 = fmt.Sprintf("%d.%d.%d.%d/32", IPv4byte1, IPv4byte2, IPv4byte3, pattern)
			break
		}
		if index+1 == len(configs) {
			ipv4 = fmt.Sprintf("%d.%d.%d.%d/32", IPv4byte1, IPv4byte2, IPv4byte3, pattern+1)
		}
		pattern++
	}
	return ipv4
}

/*
Автопоиск интерфейса и ip для конфигурации сервера.
*/
func setServerParams() (string, string) {
	out, err := exec.Command("bash", "-c", "ip r").Output()
	check(err)
	var serverIp, serverIntf string
	defaultRoute := strings.Split(string(out), " ")[:5] // первая строка "default via 192.168.11.1 dev vlan601 proto static metric 404 ..."
	ip := defaultRoute[2]
	gate4 := net.ParseIP(ip)
	serverIntf = defaultRoute[4]
	interfaces, err := net.Interfaces()
	check(err)
	for _, interf := range interfaces {
		// Список адресов для каждого сетевого интерфейса
		addrs, err := interf.Addrs()
		check(err)
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
	fmt.Println("Generate keys...")
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
func showPeers() {
	out, err := exec.Command("bash", "-c", "sudo wg show").Output()
	check(err)
	fmt.Println(string(out))
}

/*
Управление службой wg-quick.
*/
func commandServer(cmd string) {
	server := readServerConfigFile()
	command := fmt.Sprintf("sudo systemctl %s wg-quick@%s.service", cmd, server.Alias)
	out, err := exec.Command("bash", "-c", command).Output()
	check(err)
	fmt.Println(string(out))
}
