package main

const SERVER_TEMPLATE = `[Interface]
PrivateKey = {{ .ServerPrivateKey}}
Address = {{ .LocalAddress}}
ListenPort = {{ .ListenPort}}
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o {{ .Eth}} -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o {{ .Eth}} -j MASQUERADE
{{ range .Users}}
[Peer]
# Name = {{ .Name}}
PublicKey = {{ .ClientPublicKey}}
AllowedIPs = {{ .ClientLocalAddress}}
{{ end}}
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
gwg instal	- установка wireguard-сервера.
gwg show	- просмотр состояния wireguard-сервера.
gwg add		- добавления пользователя.
gwg remove	- удаление пользователя.
gwg stat	- просмотр подробной статистики wireguard-сервера. 
`
