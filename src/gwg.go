package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/PavelMilanov/go-wg-manager/tc"
)

var VERSION string

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(server.MENU)
		return nil
	}

	switch args[0] {
	case "init":
		return server.ConfigureSystem()
	case "show":
		return server.ShowPeers()
	case "stat":
		return server.ReadWgDump()
	case "install":
		set := flag.NewFlagSet("install", flag.ContinueOnError)
		alias := set.String("name", "wg0", "название сервера")
		network := set.String("network", "10.0.0.1/24", "приватный адрес сервера")
		port := set.Int("port", 51830, "порт сервера")
		if err := parseFlags(set, args[1:]); err != nil {
			return err
		}
		return server.InstallServer(*alias, *network, *port)
	case "add", "remove", "block", "unblock":
		set := flag.NewFlagSet(args[0], flag.ContinueOnError)
		alias := set.String("name", "", "имя пользователя")
		if err := parseFlags(set, args[1:]); err != nil {
			return err
		}
		switch args[0] {
		case "add":
			return server.AddUser(*alias)
		case "remove":
			return server.RemoveUser(*alias)
		case "block":
			return server.ChangeStatusUser(*alias, "block")
		default:
			return server.ChangeStatusUser(*alias, "unblock")
		}
	case "version":
		fmt.Printf("gwg version: %s\ngo version: %s\n", VERSION, runtime.Version())
		return nil
	case "tc":
		return runTC(args[1:])
	default:
		fmt.Print(server.MENU)
		return nil
	}
}

func runTC(args []string) error {
	if len(args) == 0 {
		fmt.Print(tc.TC_DEFAULT_MENU)
		return nil
	}
	switch args[0] {
	case "service":
		if len(args) < 2 {
			fmt.Print(tc.TC_SERVICE_DEFAULT_MENU)
			return nil
		}
		switch args[1] {
		case "up":
			set := flag.NewFlagSet("up", flag.ContinueOnError)
			speed := set.String("s", "", "скорость")
			fullSpeed := set.String("ms", "", "максимальная скорость")
			if err := parseFlags(set, args[2:]); err != nil {
				return err
			}
			return tc.UpService(*speed, *fullSpeed)
		case "down":
			return tc.DownService()
		case "restart":
			return tc.RestartService()
		case "show":
			return tc.ShowService()
		default:
			fmt.Print(tc.TC_SERVICE_DEFAULT_MENU)
			return nil
		}
	case "bw":
		if len(args) < 2 {
			fmt.Print(tc.TC_BW_DEFAULT_MENU)
			return nil
		}
		switch args[1] {
		case "add":
			set := flag.NewFlagSet("add", flag.ContinueOnError)
			description := set.String("d", "default", "описание")
			minimum := set.String("m", "50Mbit", "минимальная скорость")
			ceiling := set.String("c", "950Mbit", "допустимая скорость")
			if err := parseFlags(set, args[2:]); err != nil {
				return err
			}
			return tc.AddBandwidth(*description, *minimum, *ceiling)
		case "remove":
			set := flag.NewFlagSet("remove", flag.ContinueOnError)
			class := set.String("id", "", "id класса")
			if err := parseFlags(set, args[2:]); err != nil {
				return err
			}
			return tc.RemoveBandwidth(*class)
		case "show":
			return tc.ShowBandwidth()
		default:
			fmt.Print(tc.TC_BW_DEFAULT_MENU)
			return nil
		}
	case "ft":
		if len(args) < 2 {
			fmt.Print(tc.TC_FT_DEFAULT_MENU)
			return nil
		}
		switch args[1] {
		case "add":
			set := flag.NewFlagSet("add", flag.ContinueOnError)
			description := set.String("d", "", "описание")
			user := set.String("u", "", "имя пользователя")
			class := set.String("c", "1", "класс")
			if err := parseFlags(set, args[2:]); err != nil {
				return err
			}
			return tc.AddFilter(*description, *user, *class)
		case "remove":
			set := flag.NewFlagSet("remove", flag.ContinueOnError)
			filter := set.String("id", "", "id фильтра")
			if err := parseFlags(set, args[2:]); err != nil {
				return err
			}
			return tc.RemoveFilter(*filter)
		case "show":
			return tc.ShowFilter()
		default:
			fmt.Print(tc.TC_FT_DEFAULT_MENU)
			return nil
		}
	default:
		fmt.Print(tc.TC_DEFAULT_MENU)
		return nil
	}
}

func parseFlags(set *flag.FlagSet, args []string) error {
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", set.Args())
	}
	return nil
}
