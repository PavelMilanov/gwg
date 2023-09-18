package main

import (
	"fmt"
	"os"
)

func addUSer() {
	var alias string
	fmt.Println("Enter client description:")
	alias_value, _ := fmt.Scanf("%s", &alias)
	if alias_value == 0 {
		os.Exit(1)
	}
	clientPrivKey, clientPubKey := generateKeys()
	fmt.Println(clientPubKey)
	readConfigFile(WG_MANAGER_DIR)
	config := UserConfig{
		ClientPrivateKey:   clientPrivKey,
		ClientLocalAddress: "",
		ServerPublicKey:    "",
		ServerIp:           "",
		ServerPort:         1,
	}
	config.addConfigUser(alias)
}
