package command

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// reset kubernentes cluster
func ResetKubernentesCluster(cobraFlag FlagCobra) {
	cobraFlag.PrintFlagCobraConfig()
	taskResetCommand := TaskResetCommand()
	resetCommandTask := &ResetCommandTask{}
	if cobraFlag.ListTask {
		flagListTaskCommandExec(resetCommandTask)
	} else {
		log.Printf("Program starts host-level kubernentes install, Please wait a moment.\n")
		TaskCommandExec("reset", cobraFlag, taskResetCommand)
		log.Println("Program tidbkube command init exec success.")
	}
}

// ResetCommandTask struct show machine ssh shell reset command,mainly used for reset kubernentes cluster
type ResetCommandTask struct {
	ExecKubeadmReset         string `json:"exec_kubeadm_reset"`
	CleanKubeletService      string `json:"clean_kubelet_service"`
	StopNodeService          string `json:"stop_node_service"`
	CleanKubernentesDir      string `json:"clean_kubernentes_dir"`
	CleanKubeComponentConfig string `json:"clean_kube_component_config"`
	CleanKubeOtherConfig     string `json:"clean_kube_other_config"`
	CleanIpvsRule            string `json:"clean_ipvs_rule"`
	CleanHomeKubeDir         string `json:"clean_home_kube_dir"`
	CleanEtcHostFile         string `json:"clean_etc_host_file"`
	CleanKubeletServiceConf  string `json:"clean_kubelet_service_conf"`
	CleanKubeadm10Conf       string `json:"clean_kubeadm_10_conf"`
	CleanKubeletConfig       string `json:"clean_kubelet_config"`
	SystemDaemonReload       string `json:"system_daemon_reload"`
	CleanKubeComponent       string `json:"clean_kube_component"`
	CleanDockerImages        string `json:"clean_docker_images"`
	CleanInstallPkg          string `json:"clean_install_pkg"`
	CleanKubeAll             string `json:"clean_kube_all"`
}

// TaskResetCommand function
func TaskResetCommand() *ResetCommandTask {
	return &ResetCommandTask{
		ExecKubeadmReset:    "kubeadm reset --force",
		CleanKubeletService: `systemctl stop kubelet && systemctl disable kubelet`,
		StopNodeService:     `netstat -anlp|grep -E '10250|10251|10252|2379|2480' | grep -v '-' | awk '{print $7}' | grep '/' | awk -F '/' '{print $1}' | xargs  -r kill -9 && echo true`,
		CleanKubernentesDir: `rm -rf /etc/kubernetes/manifests && rm -rf  /etc/kubernetes/pki`,
		CleanKubeComponentConfig: `rm -rf /etc/kubernetes/admin.conf && rm -rf /etc/kubernetes/kubelet.conf && rm -rf  /etc/kubernetes/bootstrap-kubelet.conf && rm -rf /etc/kubernetes/controller-manager.conf && rm -rf /etc/kubernetes/scheduler.conf
`,
		CleanKubeOtherConfig:    `rm -rf /var/lib/etcd && rm -rf /var/lib/kubelet && rm -rf /etc/cni/net.d && rm -rf /var/lib/dockershim && rm -rf /var/run/kubernetes && rm -rf /var/lib/cni`,
		CleanIpvsRule:           "ipvsadm --clear",
		CleanHomeKubeDir:        "rm -rf $HOME/.kube",
		CleanEtcHostFile:        `sed -i '/apiserver.cluster.local/d' /etc/hosts`,
		CleanKubeletServiceConf: `rm -rf /etc/systemd/system/kubelet.service`,
		CleanKubeadm10Conf:      `rm -rf /etc/systemd/system/kubelet.service.d/10-kubeadm.conf`,
		CleanKubeletConfig:      `rm -rf /var/lib/kubelet/config.yaml`,
		SystemDaemonReload:      `systemctl daemon-reload`,
		CleanKubeAll:            "",
		CleanKubeComponent: `rm -rf /usr/bin/kubeadm && rm -rf /usr/bin/kubectl && rm -rf /usr/bin/kubelet && rm -rf
/usr/bin/lvscare`,
		CleanDockerImages: `docker images|awk 'NR>1{print $3}'| xargs docker rmi`,
		CleanInstallPkg:   "",
	}

}

func CleanKubeAll(cobraCmd string, kubeInstaller *KubeInstaller, cobraFlag FlagCobra) {
	if cobraFlag.PkgPath == "" {
		log.Println("Program tidbkube clean install pkg need set flag --pkg-path (install offline path)")
		os.Exit(1)
	}
	// remove docker images
	removeDockerImageCmd := fmt.Sprintf(`docker rm $(sudo docker ps -qf status=exited) && docker images |awk 'NR>1{print $1":"$2"\t
"$3}'|awk '{print $2}'|xargs docker rmi`)
	ExecTaskCommandMain(cobraCmd, "clean_docker_images", kubeInstaller.Hosts, removeDockerImageCmd, cobraFlag)
	// remove kube component
	removeKubeComponentCmd := fmt.Sprintf(`rm -rf /usr/bin/kubeadm && rm -rf /usr/bin/kubectl && rm -rf /usr/bin/kubelet && rm -rf
/usr/bin/lvscare`)
	ExecTaskCommandMain(cobraCmd, "clean_kube_component", kubeInstaller.Hosts, removeKubeComponentCmd, cobraFlag)

	// remove kube pkg
	baseDir := filepath.Dir(cobraFlag.PkgPath)
	removePkgCmd := fmt.Sprintf("rm -rf %s/kube && rm -rf %s", baseDir, cobraFlag.PkgPath)
	ExecTaskCommandMain(cobraCmd, "clean_install_pkg", kubeInstaller.Hosts, removePkgCmd, cobraFlag)

}
