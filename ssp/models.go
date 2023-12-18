package ssp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/PavelMilanov/go-wg-manager/paths"
)

type TunService struct {
	SSInt      string
	SSIP       string
	SSPort     int
	SSPassword string
}

/*
Генерирует json файл конфигурации ssp.
*/
func (config *TunService) createJsonFIle() {
	file, _ := json.MarshalIndent(config, "", "\t")
	filename := fmt.Sprintf("%s/%s", paths.SSP_DIR, paths.SSP_CONFIG_FILE)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Configuration added successfully")
}

/*
Генерирует исполняемый файл конфигурации ssp.
*/
func (config *TunService) createConfigFile() {
	configFile := fmt.Sprintf("%s/%s", paths.SSP_DIR, paths.SSP_CONFIG_FILE)
	templ, err := template.New("ssp").Parse(TUN_TEMPLATE)
	file, err := os.OpenFile(configFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0751)
	err = templ.Execute(file, config)
	if err != nil {
		fmt.Println("Error writing ssp config")
		fmt.Println(err)
		os.Remove(configFile)
	}
	defer file.Close()
}

/*
Генерирует файл службы tc и запускает ее.
*/
func (config *TunService) createService() {
	configFile := fmt.Sprintf("%s/%s", paths.SSP_DIR, paths.SSP_SERVICE_FILE)
	templ, err := template.New("ssp").Parse(TUN_SERVICE)
	file, err := os.OpenFile(configFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0751)
	err = templ.Execute(file, config)
	if err != nil {
		fmt.Println("Error writing ssp config")
		fmt.Println(err)
		os.Remove(configFile)
	}
	defer file.Close()
	copy := fmt.Sprintf("sudo mv %s /etc/systemd/system/", paths.SSP_SERVICE_FILE)
	cmd0 := exec.Command("bash", "-c", copy)
	cmd0.Stdout = os.Stdout
	cmd0.Stdin = os.Stdin
	cmd0.Stderr = os.Stderr
	cmd0.Run()
	enable := fmt.Sprintf("sudo systemctl enable tc.service")
	cmd := exec.Command("bash", "-c", enable)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	fmt.Println("Ssp started successfully")
}

func (config *TunService) stopService() {
	stop := fmt.Sprintf("sudo systemctl disable %s", paths.SSP_SERVICE_FILE)
	cmd := exec.Command("bash", "-c", stop)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	filename := fmt.Sprintf("/etc/systemd/system/%s", paths.SSP_SERVICE_FILE)
	os.Remove(filename)
}
