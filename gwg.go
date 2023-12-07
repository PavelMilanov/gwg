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
			fmt.Println("gwg version: 0.2.6")
		case "tc":
			if len(os.Args) > 2 {
				switch os.Args[2] {
				case "service":
					if len(os.Args) > 3 {
						switch os.Args[3] {
						case "up":
							// _, intf := setServerParams()
							tc.UpService("vlan601", "2mbit", "2mbit")
							// tc.UpService(intf)
						case "down":
							tc.DownService()
						case "show":
							tc.ShowService()
						}
					} else {
						fmt.Print(tc.TC_DEFAULT_MENU)
					}
				case "bw":
					if len(os.Args) > 3 {
						switch os.Args[3] {
						case "add":
							bw := flag.NewFlagSet("add", flag.ExitOnError)
							description := bw.String("d", "", "описание")
							min := bw.String("m", "50Mbit", "минимальная скорость")
							ceil := bw.String("c", "950Mbit", "допустимая скорость")
							bw.Parse(os.Args[4:])
							tc.AddBandwidth(*description, *min, *ceil)
						case "remove":
							bw := flag.NewFlagSet("remove", flag.ExitOnError)
							class := bw.String("id", "", "id класса")
							bw.Parse(os.Args[4:])
							tc.RemoveBandwidth(*class)
						case "show":
							tc.ShowBandwidth()
						default:
							fmt.Print(tc.TC_BW_DEFAULT_MENU)
						}
					} else {
						fmt.Print(tc.TC_BW_DEFAULT_MENU)
					}
				case "ft":
					if len(os.Args) > 3 {
						switch os.Args[3] {
						case "add":
							bw := flag.NewFlagSet("add", flag.ExitOnError)
							description := bw.String("d", "", "описание")
							user := bw.String("u", "", "имя пользователя")
							class := bw.String("c", "1", "полоса пропускания")
							bw.Parse(os.Args[4:])
							tc.AddFilter(*description, *user, *class)
						case "remove":
							bw := flag.NewFlagSet("remove", flag.ExitOnError)
							filter := bw.String("id", "", "id фильтра")
							bw.Parse(os.Args[4:])
							tc.RemoveFilter(*filter)
						case "show":
							tc.ShowFilter()
						default:
							fmt.Print(tc.TC_FT_DEFAULT_MENU)
						}
					} else {
						fmt.Print(tc.TC_FT_DEFAULT_MENU)
					}
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
