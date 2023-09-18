package main

import (
	"fmt"
	"os"
)

const (
	SERVER_DIR       = "/etc/wireguard/"
	WG_MANAGER_DIR   = SERVER_DIR + ".wg_manager"
	USERS_CONFIG_DIR = WG_MANAGER_DIR + ".configs"
)

func main() {
	switch os.Args[1] {
	case "config":
		configureServer("test", "test1")
	case "install":
		installServer()
	case "show":
		fmt.Println("show wg interfases")
	case "add":
		addUSer()
	case "remove":
		fmt.Println("remove user")
	}
}
