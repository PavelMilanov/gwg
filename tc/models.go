package tc

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PavelMilanov/go-wg-manager/paths"
)

type TcConfig struct {
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
