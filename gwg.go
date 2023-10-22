package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Printf(MENU)
	// 	}
	// }()
	switch os.Args[1] {
	case "config":
		configureServer("private", "publick", "wg0", "10.0.0.1/24", 51830) // for dev
	case "init":
		configureSystem()
	case "show":
		showPeers()
	case "stat":
		readWgDump()
	case "install":
		initSystem()
		installCommand := flag.NewFlagSet("install", flag.ExitOnError)
		alias := installCommand.String("name", "wg0", "название сервера")
		network := installCommand.String("network", "10.0.0.1/24", "приватный адрес сервера")
		port := installCommand.Int("port", 51830, "порт сервера")
		installCommand.Parse(os.Args[2:])
		installServer(*alias, *network, *port)
	case "add":
		addCommand := flag.NewFlagSet("add", flag.ExitOnError)
		alias := addCommand.String("name", "", "имя пользователя")
		addCommand.Parse(os.Args[2:])
		addUSer(*alias)
	case "remove":
		removeCommand := flag.NewFlagSet("remove", flag.ExitOnError)
		alias := removeCommand.String("name", "", "имя пользователя")
		removeCommand.Parse(os.Args[2:])
		removeUser(*alias)
	case "block":
		blockCommand := flag.NewFlagSet("block", flag.ExitOnError)
		alias := blockCommand.String("name", "", "имя пользователя")
		blockCommand.Parse(os.Args[2:])
		changeStatusUser(*alias, "block")
	case "unblock":
		unblockCommand := flag.NewFlagSet("unblock", flag.ExitOnError)
		alias := unblockCommand.String("name", "", "имя пользователя")
		unblockCommand.Parse(os.Args[2:])
		changeStatusUser(*alias, "unblock")
	case "version":
		fmt.Println("gwg version: 0.2.5")
	// case "test":
	// 	configureSystem()
	default:
		fmt.Print(MENU)
	}
}
