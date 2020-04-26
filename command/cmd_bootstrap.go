package command

import (
	"log"
)

// BootstrapSystemEnvironment function
func BootstrapSystemEnvironment(cobraFlag FlagCobra) {
	cobraFlag.PrintFlagCobraConfig()
	taskBootstrapCommand := TaskBootstrapCommand()
	if cobraFlag.ListTask {
		flagListTaskCommandExec(taskBootstrapCommand)
	} else {
		log.Printf("Program starts host-level environment bootstrap, Please wait a moment.\n")
		TaskCommandExec("bootstrap", cobraFlag, taskBootstrapCommand)
		log.Printf("Program host-level environment bootstrap Done, Please Execute the next round.\n")
	}

}

// BootstrapCommandTask struct show machine ssh shell bootstrap command,mainly used for bootstrap host system env
type BootstrapCommandTask struct {
	//BootstrapSystemPackage    string `json:"bootstrap_system_package"`
	BootstrapOpenIVPS              string `json:"bootstrap_open_ivps"`
	BootstrapEnableDocker          string `json:"bootstrap_enable_docker"`
	BootstrapDisableNetworkManager string `json:"bootstrap_disable_network_manager"`
	BootstrapDockerDaemonConfig    string `json:"bootstrap_docker_daemon_config"`
	BootstrapChronyServer          string `json:"bootstrap_chrony_server"`
	StartChronyServer              string `json:"start_chrony_server"`
	BootstrapDisableFirewalld      string `json:"bootstrap_disable_firewalld"`
	BootstrapDisableSelinux        string `json:"bootstrap_disable_selinux"`
	BootstrapKernelParams          string `json:"bootstrap_kernel_params"`
	BootstrapKernelProfile         string `json:"bootstrap_kernel_profile"`
	BootstrapDisableTHP            string `json:"bootstrap_disable_thp"`
	BootstrapSecurityParams        string `json:"bootstrap_security_params"`
	BootstrapCpupowerSet           string `json:"bootstrap_cpupower_set"`
}

// TaskBootstrapCommand function
func TaskBootstrapCommand() *BootstrapCommandTask {
	return &BootstrapCommandTask{
		//BootstrapSystemPackage: "yum install -y conntrack-tools libseccomp libtool-ltdl keepalived haproxy chrony*",
		BootstrapOpenIVPS:              `modprobe -- ip_vs && modprobe -- ip_vs_rr && modprobe -- ip_vs_wrr && modprobe -- ip_vs_sh && modprobe -- nf_conntrack_ipv4`,
		BootstrapDisableNetworkManager: `systemctl stop NetworkManager && systemctl disable NetworkManager`,
		BootstrapDockerDaemonConfig: `
cat > /etc/docker/daemon.json <<EOF
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ]
}
EOF`,
		BootstrapEnableDocker:     `if [ ! -d /etc/systemd/system/docker.service.d ]; then mkdir -p /etc/systemd/system/docker.service.d; fi && systemctl daemon-reload && systemctl restart docker && systemctl enable docker`,
		BootstrapChronyServer:     `sed -i 's/^.*centos.pool.ntp.org/#&/g' /etc/chrony.conf && sed -i 's/^#\(.*centos.pool.ntp.org\)/\1/' /etc/chrony.conf`,
		StartChronyServer:         `systemctl start chronyd.service && systemctl enable chronyd.service`,
		BootstrapDisableFirewalld: "systemctl stop firewalld && systemctl disable firewalld",
		BootstrapDisableSelinux:   `grep 'SELINUX=disabled' /etc/selinux/config || sed -i '/SELINUX/s/enforcing/disabled/' /etc/selinux/config && setenforce 0 || echo "yes"`,
		BootstrapKernelParams: `
cat << EOF >> /etc/sysctl.conf
net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
net.ipv4.ip_forward=1
net.ipv4.tcp_tw_recycle=0
vm.swappiness=0
vm.overcommit_memory=1
vm.panic_on_oom=0
fs.inotify.max_user_instances=8192
fs.inotify.max_user_watches=1048576
fs.file-max=52706963
fs.nr_open=52706963
net.ipv6.conf.all.disable_ipv6=1
net.netfilter.nf_conntrack_max=2310720
net.core.somaxconn = 32768
net.ipv4.tcp_syncookies = 0
EOF`,
		BootstrapKernelProfile: "sysctl -p /etc/sysctl.conf && modprobe br_netfilter",
		BootstrapDisableTHP:    "echo never > /sys/kernel/mm/transparent_hugepage/enabled && echo never > /sys/kernel/mm/transparent_hugepage/defrag",
		BootstrapSecurityParams: `
cat << EOF >> /etc/security/limits.conf
* soft nofile 1000000
* hard nofile 1000000
* soft stack  10240
* soft nproc 65536
* hard nproc 65536
* soft memlock unlimited
* hard memlock unlimited
EOF`,
		BootstrapCpupowerSet: "cpupower frequency-set --governor performance",
	}
}
