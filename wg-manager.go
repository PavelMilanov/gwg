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

func main() {
	switch os.Args[1] {
	case "config":
		configureServer("test", "test1") // for dev
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
