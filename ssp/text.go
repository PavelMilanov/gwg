package ssp

const SSP_DEFAULT_MENU = `Описание: подсистема работы gwg-сервера в режиме proxy.
Защищает от блокировок, но снижает скорость.

gwg ssp start - запуск proxy-режима.
gwg ssp stop  - остановка proxy-режима.
gwg ssp show  - просмотр текущей конфигурации.

Помощь: gwg ssp (start|stop|show) -h
`

const SSP_START_MENU = `
`

const TUN_INSTALL = `
wget https://github.com/PavelMilanov/gwg/releases/download/v0.2.6.1/ssproxy.tar

sudo tar -C /usr/bin/ -xvf ssproxy.tar
sudo sh -c "echo net.ipv4.conf.all.rp_filter=0 >> /etc/sysctl.conf"
sudo sysctl -p
`

const TUN_TEMPLATE = `
sudo ip tuntap add mode tun dev tun0
sudo ip addr add 198.168.0.1/30 dev tun0
sudo ip link set dev tun0 up

sudo ip route del default
sudo ip route add default dev tun0 metric 1
sudo ip route add default dev {{ .SSInt}} metric 10
sudo /usr/bin/tun2socks -device tun0 -proxy ss://chacha20-ietf-poly1305:{{ .SSPassword}}@{{ .SSIP}}:{{ .SSPort}} -interface {{ .SSInt}}
`

const TUN_SERVICE = `[Unit]
Description=SS-proxy
After=wg-quick@wg0.service 

[Service]
Type=simple
ExecStart=/etc/wireguard/.ssp/ssp.sh

[Install]
WantedBy=multi-user.target
`
