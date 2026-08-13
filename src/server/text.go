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
