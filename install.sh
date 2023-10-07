#!/usr/bin/bash

set -e

echo "Installing Wireguard Server..."
sudo apt install -y wireguard

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

sh -c "su - $USER"
echo "Done"
