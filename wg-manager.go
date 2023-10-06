package main

import (
	"fmt"
	"os"
)

const (
	SERVER_DIR       = "/etc/wireguard/"
	WG_MANAGER_DIR   = SERVER_DIR + ".wg_manager" // необходимо создать
	USERS_CONFIG_DIR = SERVER_DIR + ".configs"    // необходимо создать
	USERS_DIR        = SERVER_DIR + "users"       // необходимо создать
)

func init() {
	initSystem()
}

func main() {
	switch os.Args[1] {
	case "config":
		configureServer("private", "publick", "wg0") // for dev
	case "install":
		var alias string
		fmt.Println("Enter alias: 'wg0'")
		alias_value, _ := fmt.Scanf("%s\r", &alias)
		if alias_value == 0 {
			alias = "wg0"
		}
		installServer(alias)
	case "show":
		showPeers()
	case "add":
		var alias string
		fmt.Println("Enter client name:")
		alias_value, _ := fmt.Scanf("%s", &alias)
		if alias_value == 0 {
			os.Exit(1)
		}
		addUSer(alias)
		restartServer()
	case "remove":
		var alias string
		fmt.Println("Enter client name:")
		alias_value, _ := fmt.Scanf("%s", &alias)
		if alias_value == 0 {
			os.Exit(1)
		}
		removeUser(alias)
		restartServer()
	case "stat":
		readWgDump()
	}
}
