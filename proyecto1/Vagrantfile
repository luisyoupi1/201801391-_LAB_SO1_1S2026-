Vagrant.configure("2") do |config|
  config.vm.box = "generic/ubuntu2404"
  config.vm.synced_folder ".", "/proyecto", type: "rsync", rsync__auto: true

  machines = {
    "vm1" => { ip: "192.168.56.11", memory: 2048, cpus: 2 },
    "vm2" => { ip: "192.168.56.12", memory: 2048, cpus: 2 },
    "vm3" => { ip: "192.168.56.13", memory: 2048, cpus: 2 }
  }

  machines.each do |name, options|
    config.vm.define name do |machine|
      machine.vm.hostname = "so1-#{name}"
      machine.vm.network "private_network", ip: options[:ip]
      machine.vm.provider :libvirt do |libvirt|
        libvirt.cpus = options[:cpus]
        libvirt.memory = options[:memory]
      end
      machine.vm.provision "shell", path: "scripts/vm-common.sh"
    end
  end
end
