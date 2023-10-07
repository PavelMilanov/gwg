package main

import (
	"fmt"
	"os"
)

func init() {
	initSystem()
}

func main() {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Printf(MENU)
	// 	}
	// }()
	switch os.Args[1] {
	case "config":
		configureServer("private", "publick", "wg0") // for dev
	case "install":
		var alias string
		fmt.Println(string(YELLOW), "Enter alias: 'wg0'")
		alias_value, _ := fmt.Scanf("%s\r", &alias)
		if alias_value == 0 {
			alias = "wg0"
		}
		installServer(alias)
	case "show":
		showPeers()
	case "add":
		var alias string
		fmt.Println(string(YELLOW), "Enter client name:")
		alias_value, _ := fmt.Scanf("%s", &alias)
		if alias_value == 0 {
			os.Exit(1)
		}
		addUSer(alias)
	case "remove":
		var alias string
		fmt.Println(string(YELLOW), "Enter client name:")
		alias_value, _ := fmt.Scanf("%s", &alias)
		if alias_value == 0 {
			os.Exit(1)
		}
		removeUser(alias)
	case "stat":
		readWgDump()
	default:
		fmt.Print(MENU)
	}
}
