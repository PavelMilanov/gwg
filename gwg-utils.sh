#!/usr/bin/bash

set -e

command=$1

function installGwg {
    echo "Installing Wireguard Server..."
    sudo apt install -y wireguard

    echo "Preparing system..."
    sudo groupadd gwg-manager
    sudo usermod -aG gwg-manager $USER
    sudo chown root:gwg-manager /etc/wireguard
    sudo chmod ug+rwx /etc/wireguard

    echo "Creating template files..."
    mv wg_template.conf /etc/wireguard/.wg_manager/ && mv client_template.conf /etc/wireguard/.wg_manager/

    echo "Set gwg PATH..."
    sudo sh -c "echo export PATH=$PATH:/usr/bin/gwg >> /etc/profile"
    source /etc/profile

    echo "Enable ipv4 forwarding..."
    sudo sh -c "echo net.ipv4.ip_forward=1 >> /etc/sysctl.conf"
    sudo sysctl -p

    echo "Install gwg-manager..."
    sudo mv gwg /usr/bin
    echo "Done"

    gwg version

    read -p 'You must log out to complete the installation. Ready [Y] ?' user
    echo
    sudo pkill -9 -u $USER

}

function updateGwg {
    curl -O https://github.com/PavelMilanov/go-wg-manager/releases/tag/latest/gwg.tar
    tar -xvzf gwg.tar
    sudo mv gwg /usr/bin
    rm gwg.tar
    mv wg_template.conf /etc/wireguard/.wg_manager/ && mv client_template.conf /etc/wireguard/.wg_manager/
}

case "$command" in
    install)
        installGwg;;
    update)
        updateGwg;;
esac
