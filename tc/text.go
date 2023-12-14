package tc

const TC_DEFAULT_MENU = `Описание: подсистема классификации трафика по разрешенной полосе пропускания.

gwg tc service - управление службой gwg traffic control.
gwg tc bw      - управление классами gwg traffic control.
gwg tc ft      - управление фильтрами gwg traffic control.

Помощь: gwg tc (service|bw|ft) -h
`

const TC_SERVICE_DEFAULT_MENU = `Описание: tc service - управление службой gwg trafic control.

gwg tc service up      - включить службу (по-умолчанию выключена).
gwg tc service down    - выключииь службу.
gwg tc service restart - перечитать конфигурацию и перезапустить службу.
gwg tc service show    - посмотреть текущую конфигурацию службы.

Помощь: gwg tc service (up|down|restart|show) -h

`

const TC_BW_DEFAULT_MENU = `Описание: tc bw - классификатор для задания ограничения скорости.

gwg tc bw add    - создать новый класс gwg traffic control.
gwg tc bw remove - удалить класс gwg traffic control.
gwg tc bw show   - просмотр существующих классов gwg traffic control.

Помощь: gwg tc bw (add|remove|show) -h

`

const TC_FT_DEFAULT_MENU = `Описание: tc ft - правила для классификации трафика по созданным классам.

gwg tc ft add    - создать новое правило gwg traffic control.
gwg tc ft remove - удалить правило gwg traffic control.
gwg tc ft show   - просмотр существующих правил gwg traffic control.

Помощь: gwg tc ft (add|remove|show) -h

`

const TC_TEMPLATE = `
sudo tc qdisc add dev {{.Intf }} root handle 1: htb default 1
sudo tc class add dev {{.Intf }} parent 1: classid 1:1 htb rate {{ .Speed}} ceil {{ .FullSpeed}} burst 15k
{{ range .Classes}}
sudo tc class add dev {{.Intf }} parent 1:1 classid 1:{{ .Class}} htb rate {{ .MinSpeed}} ceil {{ .CeilSpeed}} burst 15k
{{end}}
{{range .Filters}}
sudo tc filter add dev {{.Intf }} protocol ip parent 1:0 u32 match ip dst {{ .UserIp}} flowid 1:{{ .Class}}
{{end}}
`

const TC_SERVICE = `[Unit]
Description=Trafic Controller
After=wg-quick@wg0.service 

[Service]
Type=simple
ExecStart=/etc/wireguard/.tc/tc.sh

[Install]
WantedBy=multi-user.target
`

//sudo tc qdisc add dev wg0 root handle 1: htb default 1
//sudo tc class add dev wg0 parent 1: classid 1:1 htb rate 1000Mbit burst 15k
//sudo tc class add dev wg0 parent 1:1 classid 1:2 htb rate 2Mbit ceil 5Mbit burst 15k
//sudo tc filter add dev wg0 protocol ip parent 1:0 u32 match ip dst 10.0.0.2 flowid 1:2
