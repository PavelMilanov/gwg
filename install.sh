#!/usr/bin/bash

set -e

echo "Installing Wireguard Server..."
sudo apt install -y wireguard

echo "Creating application tools..."
mv gwg /usr/bin/
# chown -R $USER /etc/wireguard/
# mkdir /etc/wireguard/.configs && chmod -R 764 /etc/wireguard/.configs/
# mkdir /etc/wireguard/.wg_manager && chmod -R 764 /etc/wireguard/.wg_manager/
# mkdir /etc/wireguard/users && chmod -R 764 /etc/wireguard/users/
echo "export PATH=$PATH:/usr/bin/gwg"
echo "Done"