package tc

import (
	"fmt"
	"os"
	"os/exec"

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
	createTCConfig()
}

func DownService() {

}

func createTCConfig() {
	err := os.Mkdir(paths.TC_DIR, 0660)
	if err != nil {
		fmt.Println(err)
	}
}
