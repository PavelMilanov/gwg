package main

import (
	"fmt"
	"os"
)

type WgServerConfig struct {
	ServerKey  string
	Address    string
	ListenPort int
	Eth        string
}

func main() {
	switch os.Args[1] {
	case "config":
		configureServer()
	case "install":
		installServer()
	case "show":
		fmt.Println("show wg interfases")
	case "add":
		fmt.Println("add user")
	case "remove":
		fmt.Println("remove user")
	}
}
