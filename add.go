package main

import (
	"fmt"
	"os"
	"text/template"
)

func addUSer() {
	var alias string
	fmt.Println("Enter client description:")
	alias_value, _ := fmt.Scanf("%s", &alias)
	if alias_value == 0 {
		os.Exit(1)
	}
	clientPrivKey, clientPubKey := generateKeys()
	clientip := setClientIp()
	server := readServerConfigFile()
	config := UserConfig{
		ClientPrivateKey:   clientPrivKey,
		ClientPublicKey:    clientPubKey,
		ClientLocalAddress: clientip,
		ServerPublicKey:    server.ServerPublicKey,
		ServerIp:           server.PublicAddress,
		ServerPort:         server.ListenPort,
	}
	clientFile := fmt.Sprintf("%s/%s.conf", USERS_DIR, alias)
	templ, err := template.ParseFiles("./client_template.conf")
	file, err := os.OpenFile(clientFile, os.O_CREATE|os.O_WRONLY, 0666)
	err = templ.Execute(file, config)
	if err != nil {
		panic(err)
	}
	config.addConfigUser(alias)
	defer file.Close()
}

func setClientIp() string {
	configs := readClientConfigFiles()
	label := "10.0.0.2/24"
	var lastindex = 3 // так как первый ip 10.0.0.(2)
	for index, config := range configs {
		if label <= config.ClientLocalAddress {
			label = fmt.Sprintf("10.0.0.%d/24", index+2)
		}
		lastindex += index

	}
	if len(configs) > 1 && label == configs[len(configs)-1].ClientLocalAddress {
		label = fmt.Sprintf("10.0.0.%d/24", lastindex)
	}
	return label
}
