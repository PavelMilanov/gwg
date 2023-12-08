package tc

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/PavelMilanov/go-wg-manager/paths"
	"github.com/PavelMilanov/go-wg-manager/server"
)

func UpService(minSpeed string, fullSpeed string) {
	classes := readClassFile()
	filters := readFilterFile()
	tc := TcConfig{
		FullSpeed: fullSpeed,
		Classes:   classes,
		Filters:   filters,
	}
	tc.config()
	tc.generate()
	tc.createService()
}

func DownService() {

}

func RestartService() {

}

func ShowService() {

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
	config.add(configs)
}

/*
Генерирует список моделей TcClass и преобразует их в json-файл.
*/
func RemoveBandwidth(class string) {
	configs := readClassFile()
	newConfigs := []TcClass{}
	removeConfig := TcClass{}
	for _, config := range configs {
		if config.Class == class {
			removeConfig = config
			continue
		}
		newConfigs = append(newConfigs, config)
	}
	removeConfig.remove(newConfigs)
}

/*
Выводит форматированный вывод json-файла.
*/
func ShowBandwidth() {
	configs := readClassFile()
	for _, config := range configs {
		fmt.Printf("class: %s\n\tdescription: %s;\n\tmin-rate: %s;\n\tcail-rate: %s\n\n", config.Class, config.Description, config.MinSpeed, config.CeilSpeed)
	}
}

/*
Генерирует список моделей TcFilter и преобразует их в json-файл.
*/
func AddFilter(description string, userName string, classId string) {
	classes := readClassFile()
	class := TcClass{}
	users := server.ReadClientConfigFiles()
	user := server.UserConfig{}
	for _, item := range users {
		if item.Name == userName {
			user = item
		}
	}
	if (TcClass{}) == class {
		fmt.Println("Class not found. Try gwg show tc bw show")
		os.Exit(1)
	}
	for _, item := range classes {
		if item.Class == classId {
			class = item
			break
		}
	}
	if (server.UserConfig{}) == user {
		fmt.Println("User not found. Try gwg stat")
		os.Exit(1)
	}
	filters := readFilterFile()
	filter := TcFilter{
		Description: description,
		UserIp:      user.ClientLocalAddress,
		Class:       class.Class,
	}
	filters = append(filters, filter)
	filter.add(filters)
}

/*
Генерирует список моделей TcFilter и преобразует их в json-файл.
*/
func RemoveFilter(filterDesc string) {
	filters := readFilterFile()
	newFilters := []TcFilter{}
	removeFilter := TcFilter{}
	for _, filter := range filters {
		if filter.Description == filterDesc {
			removeFilter = filter
			continue
		}
		newFilters = append(newFilters, filter)
	}
	removeFilter.remove(newFilters)
}

/*
Выводит форматированный вывод json-файла.
*/
func ShowFilter() {
	filters := readFilterFile()
	for _, filter := range filters {
		fmt.Printf("filter: %s\n\tuser: %s;\n\tclass: %s;\n", filter.Description, filter.UserIp, filter.Class)
	}
}

/*
Читает файл с tc class и преобразовывает в список моделей TcClass.
*/
func readClassFile() []TcClass {
	config := []TcClass{}
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_CLASS_FILE)
	content, err := os.ReadFile(filename)
	if err != nil { // // если не было создано ни одного класса, файла еще нет
		fmt.Println("Classes not configured")
	}
	json.Unmarshal(content, &config)
	return config
}

/*
Читает файл с tc filter и преобразовывает в список моделей TcFilter.
*/
func readFilterFile() []TcFilter {
	filter := []TcFilter{}
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_FILTER_FILE)
	content, err := os.ReadFile(filename)
	if err != nil { // // если не было создано ни одного фильтра, файла еще нет
		fmt.Println("Filters not configured")
	}
	json.Unmarshal(content, &filter)
	return filter
}

/*
Читает файл tc и преобразовывает в модель TcConfig.
*/
func readTcFile() TcConfig {
	tc := TcConfig{}
	filename := fmt.Sprintf("%s/%s", paths.TC_DIR, paths.TC_FILE)
	content, err := os.ReadFile(filename)
	if err != nil { // // если не было создано ни одного фильтра, файла еще нет
		fmt.Println("tc not configured")
	}
	json.Unmarshal(content, &tc)
	return tc
}
