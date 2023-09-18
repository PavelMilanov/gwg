package main

import (
	"fmt"
	"os"
)

const SERVER_DIR = "/etc/wireguard/"
const WG_MANAGER_DIR = SERVER_DIR + ".wg_manager"

func main() {
	switch os.Args[1] {
	case "config":
		configureServer()
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
