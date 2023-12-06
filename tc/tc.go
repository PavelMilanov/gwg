package tc

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/PavelMilanov/go-wg-manager/paths"
)

func ShowService() {

}

func UpService(intIntf string, minSpeed string, ceilSpeed string) {
	command := fmt.Sprintf("sudo cat /sys/class/net/%s/speed", intIntf)
	out, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		fmt.Println(err)
	}
	fullSpeed := string(out)
	fmt.Print(fullSpeed)
	// createTCConfig()
}

func DownService() {

}

func createTCConfig(config TcConfig) {
	err := os.Mkdir(paths.TC_DIR, 0660)
	if err != nil {
		fmt.Println(err)
	}
	tcFile := fmt.Sprintf("%s/tc.sh", paths.TC_DIR)
	templ, err := template.New("tc").Parse(TC_TEMPLATE)
	file, err := os.OpenFile(tcFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0660)
	err = templ.Execute(file, config)
	if err != nil {
		fmt.Println("Eror creating tc config file")
		os.Remove(tcFile)
	}
	defer file.Close()
}

/*
Генерирует список моделей TcClass и преобразует их в json-файл.
*/
func AddBandwidth(description string, minSpeed string, ceilSpeed string) {
	configs := readClassFile()
	re := regexp.MustCompile(`[0-9]{1,3}`) //
	idx := re.FindString(minSpeed)         // 20Mbit => 20
	config := TcClass{
		Class:       idx,
		Description: description,
		MinSpeed:    minSpeed,
		CeilSpeed:   ceilSpeed,
	}
	configs = append(configs, config)
	file, _ := json.MarshalIndent(configs, "", " ")
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, "classes")
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
}

func RemoveBandwidth() {

}

/*
Выводит форматированный вывод json-файла tc/classses
*/
func ShowBandwidth() {
	configs := readClassFile()
	for _, config := range configs {
		fmt.Printf("class: %s, description: %s, min-rate: %s, cail-rate: %s", config.Class, config.Description, config.MinSpeed, config.CeilSpeed)
	}
}

/*
Читает файл с tc class и преобразовывает в список моделей TcClass.
*/
func readClassFile() []TcClass {
	config := []TcClass{}
	content, err := os.ReadFile("tc/test.json")
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(content, &config)
	fmt.Println(config)
	return config
}
