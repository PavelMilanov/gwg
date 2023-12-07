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
	UserIp string
	Class  TcClass
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
	file, _ := json.MarshalIndent(configs, "", " ")
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_CLASS_FILE)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("class: %s\n\tdescription: %s;\n\tmin-rate: %s;\n\tcail-rate: %s;\nRemoved successfully\n", class.Class, class.Description, class.MinSpeed, class.CeilSpeed)
}
