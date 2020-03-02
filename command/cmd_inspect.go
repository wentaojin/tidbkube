package command

import (
	"log"
)

// InspectSystemEnvironment function
func InspectSystemEnvironment(cobraFlag FlagCobra) {
	cobraFlag.PrintFlagCobraConfig()
	taskInspectCommand := TaskInspectCommand()
	if cobraFlag.ListTask {
		flagListTaskCommandExec(taskInspectCommand)
	} else {
		log.Printf("Program starts host-level environment check, Please wait a moment.\n")
		TaskCommandExec("inspect", cobraFlag, taskInspectCommand)
		log.Printf("Program host-level environment check Done, Please Execute the next round.\n")
	}
}

// InspectCommandTask show machine ssh shell inspect command,mainly used for inspect host system env
type InspectCommandTask struct {
	CheckSystemVersion      string `json:"check_system_version"`
	CheckSystemCPU          string `json:"check_system_cpu"`
	CheckSystemMemory       string `json:"check_system_memory"`
	CheckSystemNetwork      string `json:"check_system_network"`
	CheckSystemHostname     string `json:"check_system_hostname"`
	CheckSystemMacAddr      string `json:"check_system_mac_addr"`
	CheckSystemProductID    string `json:"check_system_product_id"`
	CheckK8sPartDefaultPort string `json:"check_k8s_part_default_port"`
	CheckSystemSwap         string `json:"check_system_swap"`
	CheckDockerInstall      string `json:"check_docker_install"`
	CheckDockerVersion      string `json:"check_docker_version"`
	CheckIpvsadmInstall     string `json:"check_ipvsadm_install"`
}

// TaskInspectCommand function
func TaskInspectCommand() *InspectCommandTask {
	return &InspectCommandTask{
		CheckSystemVersion: "cat /etc/redhat-release",
		CheckSystemCPU:     `echo $(grep "processor" /proc/cpuinfo|sort -u|wc -l)`,
		CheckSystemMemory:  "echo $(($(echo $(cat /proc/meminfo  | grep -i memtotal) | cut '-d ' -f 2) / 1024 / 1024))",
		// default value,the corresponding commands will be automatically configured according to the configuration file in the future
		CheckSystemNetwork:      "ping 127.0.0.1",
		CheckSystemHostname:     "hostname",
		CheckSystemMacAddr:      "ip link show|grep link/ether | awk '{print $2}'",
		CheckSystemProductID:    "cat /sys/class/dmi/id/product_uuid",
		CheckK8sPartDefaultPort: `netstat -anlp | awk '{print $4,$7}' |grep -Ew "2379|2480|10250|10251|10255|10252"|grep -wv '-'|sort -u|uniq`,
		CheckSystemSwap:         "cat /proc/meminfo|grep -w 'SwapTotal'|awk '{print $2}'",
		CheckDockerInstall:      "rpm -qa | grep -w 'docker'  || echo 'no'",
		CheckDockerVersion:      "docker -v |awk '{print $3}' |awk -F '.' '{print $1}'",
		CheckIpvsadmInstall:     "rpm -qa|grep ipvsadm || echo no",
	}
}
