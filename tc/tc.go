package tc

import (
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

func AddBandwidth(description string, minSpeed string, ceilSpeed string) {
	re := regexp.MustCompile(`[0-9]{1,3}`)
	idx := re.FindString(minSpeed)
	fmt.Printf(idx)
	// config := TcClass{
	// 	Class:       strconv.Atoi(idx),
	// 	Description: description,
	// 	MinSpeed:    minSpeed,
	// 	CeilSpeed:   ceilSpeed,
	// }
}

func RemoveBandwidth() {

}

func ShowBandwidth() {

}
