package main

import (
	"fmt"
	"os"
	"os/exec"
)

func installServer() {
	/*
		Основаня логика установки WG Server.
	*/
	updatePackage()
	installWgServer()
	os.Mkdir(WG_MANAGER_DIR, 0666)
	generateKeys()
}
func updatePackage() {
	/*
		Обновление пакетов deb.
	*/
	fmt.Println("Updating packages...")
	cmd := exec.Command("apt", "update", "-y")
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func installWgServer() {
	/*
		Установка пакета wireguard.
	*/
	fmt.Println("Installing WireGuard Server...")
	cmd := exec.Command("apt", "install", "-y", "wireguard")
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func generateKeys() {
	/*
		Генерация приватного и публичного ключей сервера и сохранение в файлы.
	*/
	os.Chdir(WG_MANAGER_DIR)
	fmt.Println("Generate keys...")
	cmd := exec.Command("bash", "-c", "wg genkey | tee privatekey | wg pubkey | tee publickey")
	cmd.Stderr = os.Stderr
	cmd.Run()
}
