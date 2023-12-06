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

func (class *TcClass) add() {
	file, _ := json.MarshalIndent(class, "", " ")
	filename := fmt.Sprintf("%s/%s.json", paths.TC_DIR, class.Description)
	err := os.WriteFile(filename, file, 0660)
	if err != nil {
		fmt.Println(err)
	}
}

func (class *TcClass) remove() {

}
