package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/PavelMilanov/go-wg-manager/tc"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		// case "config":
		// 	configureServer("private", "publick", "wg0", "10.0.0.1/24", 51830) // for dev
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
			fmt.Println("gwg version: 0.2.5.1")
		case "tc":
			if len(os.Args) > 2 {
				fmt.Println(os.Args[1:])
				switch os.Args[2] {
				case "show":
					tc.ShowService()
				case "up":
					// _, intf := setServerParams()
					tc.UpService("vlan601", "2mbit", "2mbit")
					// tc.UpService(intf)
				case "down":
					tc.DownService()
				default:
					fmt.Print(tc.TC_DEFAULT_MENU)
				}
			} else {
				fmt.Print(tc.TC_DEFAULT_MENU)
			}
		default:
			fmt.Print(MENU)
		}
	} else {
		fmt.Print(MENU)
	}
}
