package main

import (
	"fmt"
	"os"
	"regexp"
	"text/template"
)

func configureServer() {
	var (
		private_addr string
		port         int
		intf         string
		alias        string
	)
	fmt.Println("Enter private network: 10.0.0.1/24")
	private_addr_value, _ := fmt.Scanf("%s\n", &private_addr)
	if private_addr_value == 0 {
		private_addr = "10.0.0.1/24"
	} else {
		isValid, _ := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}/[0-9]{1,2}`, private_addr)
		if !isValid {
			fmt.Println("Enter valid value. Example: 10.0.0.1/24")
			os.Exit(1)
		}
	}
	fmt.Println("Enter listen port: 51830")
	port_value, _ := fmt.Scanf("%d\n", &port)
	if port_value == 0 {
		port = 51830
	}
	fmt.Println("Enter NAT-interface:")
	intf_value, _ := fmt.Scanf("%s\n", &intf)
	if intf_value == 0 {
		fmt.Println("Enter NAT-interface")
		os.Exit(1)
	}
	fmt.Println("Enter alias: 'wg0'")
	alias_value, _ := fmt.Scanf("%s\n", &alias)
	if alias_value == 0 {
		alias = "wg0"
	}
	config := WgServerConfig{
		ServerKey:  "test-private-key",
		Address:    private_addr,
		ListenPort: port,
		Eth:        intf,
	}
	configFile := fmt.Sprintf("%s.conf", alias)
	templ, err := template.ParseFiles("wg_template.conf")
	file, err := os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY, 0666)
	err = templ.Execute(file, config)
	if err != nil {
		panic(err)
	}
	defer file.Close()
}
