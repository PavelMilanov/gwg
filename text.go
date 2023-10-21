package main

const SERVER_TEMPLATE = `[Interface]
PrivateKey = {{ .ServerPrivateKey}}
Address = {{ .LocalAddress}}
ListenPort = {{ .ListenPort}}
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o {{ .Eth}} -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o {{ .Eth}} -j MASQUERADE
{{ range .Users}}
{{ if .Status}}[Peer]{{else}}# [Peer]{{end}}
# Name = {{ .Name }}
{{ if .Status}}PublicKey = {{ .ClientPublicKey}}{{else}}# PublicKey = {{ .ClientPublicKey}}{{end}}
{{ if .Status}}AllowedIPs = {{ .ClientLocalAddress}}{{else}}# AllowedIPs = {{ .ClientLocalAddress}}{{end}}
{{end}}
`

const CLIENT_TEMPLATE = `[Interface]
PrivateKey = {{ .ClientPrivateKey}}
Address = {{ .ClientLocalAddress}}
DNS = 8.8.8.8

[Peer]
PublicKey = {{ .ServerPublicKey}}
Endpoint = {{ .ServerIp}}:{{ .ServerPort}}
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 20
`

const MENU = `Меню утилиты gwg-manager:
gwg show	- просмотр состояния wireguard-сервера.
gwg stat	- просмотр подробной статистики wireguard-сервера. 
gwg add		- добавления пользователя.
gwg remove	- удаление пользователя.
gwg block	- блокировка пользователя.
gwg unblock - разблокировка пользователя.
`

const GWG_UTILS = `#!/usr/bin/bash

SERVER_DIR="/etc/wireguard/"
WG_MANAGER_DIR="${SERVER_DIR}.wg_manager"
USERS_CONFIG_DIR="${SERVER_DIR}.configs"
USERS_DIR="${SERVER_DIR}users"

command=$1
version=$2

function preinstallGwg {
    echo "Installing Wireguard Server..."
    sudo apt install -y wireguard iptables

    echo "Preparing system..."
    sudo groupadd gwg-manager
    sudo usermod -aG gwg-manager $USER
    sudo chown root:gwg-manager /etc/wireguard
    sudo chmod ug+rwx /etc/wireguard

    echo "Set gwg PATH..."
    sudo sh -c "echo export PATH=$PATH:/usr/bin/gwg >> /etc/profile"
    source /etc/profile

    echo "Enable ipv4 forwarding..."
    sudo sh -c "echo net.ipv4.ip_forward=1 >> /etc/sysctl.conf"
    sudo sysctl -p
	if [ $? -eq 0 ]
	then
		su - $USER ./gwg-utils.sh postinstall
	else
		echo "Error" >&2
		exit 1
	fi
}

function postinstallGwg {
    echo "Creating gwg directory..."
    mkdir $WG_MANAGER_DIR
    mkdir $USERS_CONFIG_DIR
    mkdir $USERS_DIR
	gwg version
    echo "Installing wg server..."
    gwg wireguard

    read -p 'You must log out to complete the installation. Ready [Y] ?' user
    echo
    sudo pkill -9 -u $USER
}

function updateGwg {
    wget https://github.com/PavelMilanov/go-wg-manager/releases/download/${version}/gwg.tar
    tar -xzf gwg.tar
    sudo mv gwg /usr/bin
    rm gwg.tar
}

case "$command" in
    install)
        preinstallGwg;;
    postinstall)
        postinstallGwg;;
	update)
        updateGwg;;
esac
`
