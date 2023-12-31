package tc

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/PavelMilanov/go-wg-manager/paths"
)

type TcConfig struct {
	Intf      string
	Speed     string
	FullSpeed string
	Classes   []TcClass
	Filters   []TcFilter
}

type TcClass struct {
	Class       string
	Description string
	CeilSpeed   string
	MinSpeed    string
}

type TcFilter struct {
	Description string
	UserIp      string
	Class       string
}

/*
Добавляет модель в конфигурационный файл.
*/
func (class *TcClass) add(configs []TcClass) {
	file, _ := json.MarshalIndent(configs, "", "\t")
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_CLASS_FILE)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("class: %s\n\tdescription: %s;\n\tmin-rate: %s;\n\tcail-rate: %s;\nAdded successfully\n", class.Class, class.Description, class.MinSpeed, class.CeilSpeed)
}

/*
Удаляет модель из конфигурационного файла.
*/
func (class *TcClass) remove(configs []TcClass) {
	file, _ := json.MarshalIndent(configs, "", "\t")
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_CLASS_FILE)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("class: %s\n\tdescription: %s;\n\tmin-rate: %s;\n\tcail-rate: %s;\nRemoved successfully\n", class.Class, class.Description, class.MinSpeed, class.CeilSpeed)
}

/*
Добавляет модель в конфигурационный файл.
*/
func (filter *TcFilter) add(filters []TcFilter) {
	file, _ := json.MarshalIndent(filters, "", "\t")
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_FILTER_FILE)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("filter: %s\n\tuser: %s;\n\tclass: %s;\nAdded successfully\n", filter.Description, filter.UserIp, filter.Class)
}

/*
Удаляет модель из конфигурационного файла.
*/
func (filter *TcFilter) remove(filters []TcFilter) {
	file, _ := json.MarshalIndent(filters, "", "\t")
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_FILTER_FILE)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("filter: %s\n\tuser: %s;\n\tclass: %s;\nRemoved successfully\n", filter.Description, filter.UserIp, filter.Class)
}

/*
Генерирует json файл конфигурации tc.
*/
func (tc *TcConfig) config() {
	file, _ := json.MarshalIndent(tc, "", "\t")
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_FILE)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Tc config file generated successfully")
}

/*
Генерирует исполняемый файл конфигурации tc.
*/
func (tc *TcConfig) generate() {
	tcFile := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_CONFIG_FILE)
	templ, err := template.New("tc").Parse(TC_TEMPLATE)
	file, err := os.OpenFile(tcFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0751)
	err = templ.Execute(file, tc)
	if err != nil {
		fmt.Println("Eror creating tc config file")
		os.Remove(tcFile)
	}
	defer file.Close()
	fmt.Println("Tc executable file generated successfully")
}

/*
Генерирует файл службы tc и запускает ее.
*/
func (tc *TcConfig) createService() {
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_SERVICE_FILE)
	err := os.WriteFile(filename, []byte(TC_SERVICE), 0751)
	if err != nil {
		fmt.Println(err)
	}
	copy := fmt.Sprintf("sudo mv %s /etc/systemd/system/", filename)
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
}

/*
Удаление службы tc.
*/
func (tc *TcConfig) removeSerice() {
	stop := fmt.Sprintf("sudo systemctl disable %s", paths.TC_SERVICE_FILE)
	cmd := exec.Command("bash", "-c", stop)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	filename := fmt.Sprintf("/etc/systemd/system/%s", paths.TC_SERVICE_FILE)
	os.Remove(filename)
}

/*
Запуск исполняемого файла службы gwg tc.
*/
func (tc *TcConfig) start() {
	command := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_CONFIG_FILE)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

/*
Удаляет все правила tc.
*/
func (tc *TcConfig) down() {
	command := fmt.Sprintf("sudo tc qdisc del dev wg0 root")
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}
