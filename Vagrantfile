Vagrant.configure("2") do |config|
  config.vm.box = "generic/debian12"
  config.vm.box_version = "4.3.12"
  config.vm.synced_folder "./vagrant/share", "/home/vagrant/share"

  config.vm.define "server" do |server|
    server.vm.hostname = "server"
    server.vm.provision "file", source: "./src/gwg", destination: "/home/vagrant/gwg"
    server.vm.provision "shell", inline: "chmod 0755 /home/vagrant/gwg"
  end

  config.vm.define "external-1" do |server|
    server.vm.hostname = "external-1"
    server.vm.provision "shell", inline: <<-SHELL
      export DEBIAN_FRONTEND=noninteractive
      apt update
      apt install -y wireguard
    SHELL
  end

  config.vm.define "external-2" do |server|
    server.vm.hostname = "external-2"
    server.vm.provision "shell", inline: <<-SHELL
      export DEBIAN_FRONTEND=noninteractive
      apt update
      apt install -y wireguard
    SHELL
  end

  config.vm.define "client-1" do |client|
    client.vm.hostname = "client-1"
  end

  config.vm.define "client-2" do |client|
    client.vm.hostname = "client-2"
  end
end
