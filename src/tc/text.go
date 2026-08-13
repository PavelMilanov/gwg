package tc

const TC_TEMPLATE = `#!/bin/sh
set -eu

tc qdisc add dev {{.Intf}} root handle 1: htb default 1
tc class add dev {{.Intf}} parent 1: classid 1:1 htb rate {{ .Speed}} ceil {{ .FullSpeed}} burst 15k
{{ range .Classes}}
tc class add dev {{$.Intf}} parent 1:1 classid 1:{{ .Class}} htb rate {{ .MinSpeed}} ceil {{ .CeilSpeed}} burst 15k
{{end}}
{{range .Filters}}
tc filter add dev {{$.Intf}} protocol ip parent 1:0 u32 match ip dst {{ .UserIp}} flowid 1:{{ .Class}}
{{end}}
`

const TC_SERVICE_TEMPLATE = `[Unit]
Description=Trafic Controller
After=wg-quick@{{.Intf}}.service
Requires=wg-quick@{{.Intf}}.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=sh /etc/wireguard/.tc/tc.sh

[Install]
WantedBy=multi-user.target
`
