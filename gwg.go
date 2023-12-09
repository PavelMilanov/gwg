package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/PavelMilanov/go-wg-manager/tc"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			server.ConfigureSystem()
		case "show":
			server.ShowPeers()
		case "stat":
			server.ReadWgDump()
		case "install":
			installCommand := flag.NewFlagSet("install", flag.ExitOnError)
			alias := installCommand.String("name", "wg0", "название сервера")
			network := installCommand.String("network", "10.0.0.1/24", "приватный адрес сервера")
			port := installCommand.Int("port", 51830, "порт сервера")
			installCommand.Parse(os.Args[2:])
			server.InstallServer(*alias, *network, *port)
		case "add":
			addCommand := flag.NewFlagSet("add", flag.ExitOnError)
			alias := addCommand.String("name", "", "имя пользователя")
			addCommand.Parse(os.Args[2:])
			server.AddUSer(*alias)
		case "remove":
			removeCommand := flag.NewFlagSet("remove", flag.ExitOnError)
			alias := removeCommand.String("name", "", "имя пользователя")
			removeCommand.Parse(os.Args[2:])
			server.RemoveUser(*alias)
		case "block":
			blockCommand := flag.NewFlagSet("block", flag.ExitOnError)
			alias := blockCommand.String("name", "", "имя пользователя")
			blockCommand.Parse(os.Args[2:])
			server.ChangeStatusUser(*alias, "block")
		case "unblock":
			unblockCommand := flag.NewFlagSet("unblock", flag.ExitOnError)
			alias := unblockCommand.String("name", "", "имя пользователя")
			unblockCommand.Parse(os.Args[2:])
			server.ChangeStatusUser(*alias, "unblock")
		case "version":
			fmt.Println("gwg version: 0.2.6")
		case "tc":
			if len(os.Args) > 2 {
				switch os.Args[2] {
				case "service":
					if len(os.Args) > 3 {
						switch os.Args[3] {
						case "up":
							bw := flag.NewFlagSet("up", flag.ExitOnError)
							Speed := bw.String("s", "", "скорость")
							FullSpeed := bw.String("ms", "", "максимальная скорость. Обязательный.")
							bw.Parse(os.Args[4:])
							tc.UpService(*Speed, *FullSpeed)
						case "down":
							tc.DownService()
						case "restart":
							tc.RestartService()
						case "show":
							tc.ShowService()
						}
					} else {
						fmt.Print(tc.TC_SERVICE_DEFAULT_MENU)
					}
				case "bw":
					if len(os.Args) > 3 {
						switch os.Args[3] {
						case "add":
							bw := flag.NewFlagSet("add", flag.ExitOnError)
							description := bw.String("d", "default", "описание")
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
							class := bw.String("c", "1", "класс")
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
			fmt.Print(server.MENU)
		}
	} else {
		fmt.Print(server.MENU)
	}
}
