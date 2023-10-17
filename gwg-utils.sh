#!/usr/bin/bash

set -e


SERVER_DIR="/etc/wireguard/"
WG_MANAGER_DIR="${SERVER_DIR}.wg_manager"
USERS_CONFIG_DIR="${SERVER_DIR}.configs"
USERS_DIR="${SERVER_DIR}users"

command=$1
version=$2

function preinstallGwg {
    echo "Installing Wireguard Server..."
    sudo apt install -y wireguard iptables

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

    echo "Set gwg..."
    sudo mv gwg /usr/bin
    echo "Done"

    gwg version

    su - $USER ./gwg-utils.sh server_install
}

function postinstallGwg {
    echo "Creating gwg directory..."
    mkdir $WG_MANAGER_DIR
    mkdir $USERS_CONFIG_DIR
    mkdir $USERS_DIR

    echo "Installing wg server..."
    gwg install

    read -p 'You must log out to complete the installation. Ready [Y] ?' user
    echo
    sudo pkill -9 -u $USER
}

function updateGwg {
    wget https://github.com/PavelMilanov/go-wg-manager/releases/download/${version}/gwg.tar
    tar -xzf gwg.tar
    sudo mv gwg /usr/bin
    rm gwg.tar
}

case "$command" in
    install)
        preinstallGwg;;
    server_install)
        postinstallGwg;;
    update)
        updateGwg;;
esac
