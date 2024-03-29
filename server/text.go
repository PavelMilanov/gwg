package server

const SERVER_TEMPLATE = `[Interface]
PrivateKey = {{ .ServerPrivateKey}}
Address = {{ .LocalAddress}}
ListenPort = {{ .ListenPort}}
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o {{ .Eth}} -j MASQUERADE; sudo ip rule add from {{ .PublicAddress}} table main;
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o {{ .Eth}} -j MASQUERADE; sudo ip rule del from {{ .PublicAddress}} table main;
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
DNS = {{ .DnsResolv}}

[Peer]
PublicKey = {{ .ServerPublicKey}}
Endpoint = {{ .ServerIp}}:{{ .ServerPort}}
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
`

const MENU = `Описание: gwg - cli-менеджер wireguard:

gwg show    - просмотр состояния wireguard-сервера.
gwg stat    - просмотр подробной статистики wireguard-сервера. 
gwg add     - добавления пользователя.
gwg remove  - удаление пользователя.
gwg block   - блокировка пользователя.
gwg unblock - разблокировка пользователя.
gwg tc      - модуль управления трафиком. (По-умолчанию выключен).

Помощь: gwg <command> -h
`

const GWG_UTILS = `
SERVER_DIR="/etc/wireguard/"
WG_MANAGER_DIR="${SERVER_DIR}.wg_manager"
USERS_CONFIG_DIR="${SERVER_DIR}.configs"
USERS_DIR="${SERVER_DIR}users"
TC_DIR="${SERVER_DIR}.tc"

command=$1
version=$2

function preinstallGwg {
    echo "Installing Wireguard Server..."
    sudo apt install -y wireguard iptables sudo

    echo "Preparing system..."
    sudo chown root:$USER /etc/wireguard
    sudo chmod ug+rwx /etc/wireguard

    echo "Set gwg PATH..."
    sudo sh -c "echo export PATH=$PATH:/usr/bin/gwg >> /etc/profile"
    source /etc/profile

    echo "Enable ipv4 forwarding..."
    sudo sh -c "echo net.ipv4.ip_forward=1 >> /etc/sysctl.conf"
    sudo sysctl -p
	echo "Creating gwg directory..."
    mkdir $WG_MANAGER_DIR
    mkdir $USERS_CONFIG_DIR
    mkdir $USERS_DIR
    mkdir $TC_DIR
    
    echo "Installing wg server..."
    gwg install
	gwg version
}

case "$command" in
    install)
        preinstallGwg;;
esac
`
