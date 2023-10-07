#!/usr/bin/bash

set -e

echo "Installing Wireguard Server..."
sudo apt install -y wireguard


sudo groupadd gwg-manager
sudo usermod -aG gwg-manager $USER
su - $USER"
exit
# echo "Creating application tools..."
# mv gwg /usr/bin/
sudo chown -R $USER /etc/wireguard/ && chmod -R 760 /etc/wireguard/
sudo chown root:gwg-manager /etc/wireguard
sudo chmod ug+rwx /etc/wireguard
echo "export PATH=$PATH:/usr/bin/gwg"
echo "Done"