package tc

const TC_DEFAULT_MENU = `Usage:
	gwg tc show	- просмотр конфигурации службы gwg traffic control.
	gwg tc up	- включить службу gwg traffic control.
	gwg tc down	- выключить службу gwg traffic control.	
`

const TC_TEMPLATE = `sudo tc qdisc add dev wg0 root handle 1: htb default 1
sudo tc class add dev wg0 parent 1: classid 1:1 htb rate {{ .FullSpeed}} burst 15k
{{ range .Classes}}
sudo tc class add dev wg0 parent 1:1 classid 1:{{ .Class}} htb rate {{ .MinSpeed}} ceil {{ .CeilSpeed}} burst 15k
{{end}}
{{range .Filters}}
sudo tc filter add dev wg0 protocol ip parent 1:0 u32 match ip dst {{ .UserIp}} flowid 1:{{ .Class}}
{{end}}
`

//sudo tc qdisc add dev wg0 root handle 1: htb default 1
//sudo tc class add dev wg0 parent 1: classid 1:1 htb rate 1000Mbit burst 15k
//sudo tc class add dev wg0 parent 1:1 classid 1:2 htb rate 2Mbit ceil 5Mbit burst 15k
//sudo tc filter add dev wg0 protocol ip parent 1:0 u32 match ip dst 10.0.0.2 flowid 1:2
