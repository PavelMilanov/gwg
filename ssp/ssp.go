package ssp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/PavelMilanov/go-wg-manager/paths"
)

/*
Предварительная подготовка системы к запуску tun2socks.
*/
func install() {
	installFile := "./setup.sh"
	err := os.WriteFile(installFile, []byte(TUN_INSTALL), 0751)
	if err != nil {
		fmt.Println(err)
	}
	command := fmt.Sprintf("%s install", installFile)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	fmt.Println("Proxy os utils configured successfully")
	defer os.Remove(installFile)
}

/*
Запуск proxy-службы для gwg сервера.
*/
func Start(intf string, ip string, port int, pwd string) {
	command := fmt.Sprintf("sudo systemctl is-enabled %s", paths.SSP_SERVICE_FILE)
	out, _ := exec.Command("bash", "-c", command).Output()
	if string(out) == "enabled" {
		fmt.Println("Service already enabled")
		os.Exit(1)
	}
	SspConfig := readTunFile()
	if (TunService{}) == SspConfig { // если файла нет или он пустой, создаем все новое
		config := TunService{
			SSInt:      intf,
			SSIP:       ip,
			SSPort:     port,
			SSPassword: pwd,
		}
		config.createJsonFIle()
		config.createConfigFile()
		config.createService()
		fmt.Println("Proxy mode enabled")
	} else { // если конфиги есть, просто запускаем службу
		SspConfig.createConfigFile()
		SspConfig.createService()
		fmt.Println("Proxy mode enabled")
	}
}

/*
Остановка proxy-службы для gwg сервера.
*/
func Stop() {
	SspConfig := readTunFile()
	SspConfig.stopService()
	fmt.Println("Proxy mode stopped")
}

/*
Просмотр конфигурации proxy-службы для gwg сервера.
*/
func Show() {
	command := fmt.Sprintf("sudo systemctl is-enabled %s", paths.SSP_SERVICE_FILE)
	out, _ := exec.Command("bash", "-c", command).Output()
	if string(out) != "enabled" {
		fmt.Println("Service not enabled")
		os.Exit(1)
	} else {
		SspConfig := readTunFile()
		fmt.Printf("Proxy mode for server:\n\t%s;\n\t%d;\n\t%s;", SspConfig.SSIP, SspConfig.SSPort, SspConfig.SSPassword)
	}
}

/*
Читает файл конфигурации и преобразовывает в модель TunService.
*/
func readTunFile() TunService {
	config := TunService{}
	filename := fmt.Sprintf("%s/%s", paths.SSP_DIR, paths.SSP_FILE)
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Tun Error: ", err)
	}
	json.Unmarshal(content, &config)
	return config
}
