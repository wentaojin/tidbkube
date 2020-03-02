package command

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func BuildJoinKubernentesNode(cobraFlag FlagCobra) {
	if cobraFlag.ControlPlane {
		joinKubernentesMaster(cobraFlag)
	} else {
		joinKubernentesWorker(cobraFlag)
	}
}

func joinKubernentesMaster(cobraFlag FlagCobra) {
	if len(cobraFlag.JoinNodeIP) <= 0 {
		log.Println("Func joinKubernentesMaster:::Cobra command join flag --join-node Can't Null, " +
			"Please view command help (--help)")
		os.Exit(1)
	}
	cobraFlag.PrintFlagCobraConfig()
	kubeInstaller := TidbKubeInstaller(cobraFlag)
	taskJoinMasterCommand := TaskJoinMasterCommand(kubeInstaller, cobraFlag)
	joinMasterCommandTask := &JoinMasterCommandTask{}
	if cobraFlag.ListTask && cobraFlag.ControlPlane {
		flagListTaskCommandExec(joinMasterCommandTask)
	} else {
		log.Printf("Program starts host-level kubernentes install, Please wait a moment.\n")
		TaskCommandExec("join_master", cobraFlag, taskJoinMasterCommand)
		log.Println("Program tidbkube command init exec success.")
	}
}

func joinKubernentesWorker(cobraFlag FlagCobra) {
	if len(cobraFlag.JoinNodeIP) <= 0 {
		log.Println("Func joinKubernentesWorker:::Cobra command join flag --join-node Can't Null, " +
			"Please view command help (--help)")
		os.Exit(1)
	}
	cobraFlag.PrintFlagCobraConfig()
	kubeInstaller := TidbKubeInstaller(cobraFlag)
	taskJoinWorkerCommand := TaskJoinWorkerCommand(kubeInstaller, cobraFlag)
	joinWorkerCommandTask := &JoinWorkerCommandTask{}
	if cobraFlag.ListTask {
		flagListTaskCommandExec(joinWorkerCommandTask)
	} else {
		log.Printf("Program starts host-level kubernentes install, Please wait a moment.\n")
		TaskCommandExec("join_worker", cobraFlag, taskJoinWorkerCommand)
		log.Println("Program tidbkube command init exec success.")
	}
}

// JoinMasterCommandTask struct show machine ssh shell join master command,mainly used for kubernentes master node join
type JoinMasterCommandTask struct {
	// offline package init
	SendOfflinePackage    string `json:"send_offline_package"`
	UnzipOfflinePackage   string `json:"unzip_offline_package"`
	CopyKubernentesBinary string `json:"copy_kubernentes_binary"`
	LoadDockerImages      string `json:"load_docker_images"`
	// kubernentes master node env init
	KubeEnvironmentInit string `json:"kube_environment_init"`
	// kubernentes master node join
	JoinMasterEtcHostsSet         string `json:"join_master_etc_hosts_set"`
	KubeadmJoinMaster             string `json:"kubeadm_join_master"`
	SedJoinMasterEtcHosts         string `json:"sed_join_master_etc_hosts"`
	CopyKubernentesConfig         string `json:"copy_kubernentes_config"`
	AddWorkerNodeIPVSRule         string `json:"add_worker_node_ipvs_rule"`
	UpdateWorkerIPVSStaticPodYaml string `json:"update_worker_ipvs_static_pod_yaml"`
}

func TaskJoinMasterCommand(kubeInstaller *KubeInstaller, cobraFlag FlagCobra) *JoinMasterCommandTask {
	baseDir := filepath.Dir(cobraFlag.PkgPath)
	fileName := filepath.Base(cobraFlag.PkgPath)
	unzipPkgCmd := fmt.Sprintf("cd %s && tar zxvf %s", baseDir, fileName)
	copyKubeBinary := fmt.Sprintf("cp  %s/kube/bin/* /usr/bin", baseDir)
	loadDockerImages := fmt.Sprintf("docker load -i %s/kube/images/images.tar", baseDir)

	joinMasterEtcHostCmd := kubeInstaller.joinMasterEtcHost(cobraFlag)
	copyK8sConfCmd := kubeInstaller.joinMasterConfig(cobraFlag)

	return &JoinMasterCommandTask{
		SendOfflinePackage:            "",
		UnzipOfflinePackage:           unzipPkgCmd,
		CopyKubernentesBinary:         copyKubeBinary,
		LoadDockerImages:              loadDockerImages,
		KubeEnvironmentInit:           "",
		JoinMasterEtcHostsSet:         joinMasterEtcHostCmd,
		KubeadmJoinMaster:             "",
		SedJoinMasterEtcHosts:         "",
		CopyKubernentesConfig:         copyK8sConfCmd,
		AddWorkerNodeIPVSRule:         "",
		UpdateWorkerIPVSStaticPodYaml: "",
	}
}

// JoinWorkerCommandTask struct show machine ssh shell join worker command,mainly used for kubernentes worker node join
type JoinWorkerCommandTask struct {
	// offline package init
	SendOfflinePackage    string `json:"send_offline_package"`
	UnzipOfflinePackage   string `json:"unzip_offline_package"`
	CopyKubernentesBinary string `json:"copy_kubernentes_binary"`
	LoadDockerImages      string `json:"load_docker_images"`
	// kubernentes worker node env init
	KubeEnvironmentInit string `json:"kube_environment_init"`
	// kubernentes worker node join
	JoinWorkerEtcHostsSet string `json:"join_worker_etc_hosts_set"`
	WorkerStaticPodCreate string `json:"worker_static_pod_create"`
	KubeadmJoinWorker     string `json:"kubeadm_join_worker"`
}

// TaskJoinWorkerCommand function
func TaskJoinWorkerCommand(kubeInstaller *KubeInstaller, cobraFlag FlagCobra) *JoinWorkerCommandTask {
	baseDir := filepath.Dir(cobraFlag.PkgPath)
	fileName := filepath.Base(cobraFlag.PkgPath)
	unzipPkgCmd := fmt.Sprintf("cd %s && tar zxvf %s", baseDir, fileName)
	copyKubeBinary := fmt.Sprintf("cp  %s/kube/bin/* /usr/bin", baseDir)
	loadDockerImages := fmt.Sprintf("docker load -i %s/kube/images/images.tar", baseDir)

	echoWorkerHostCmd := kubeInstaller.joinWorkerEtcHost(cobraFlag)

	return &JoinWorkerCommandTask{
		SendOfflinePackage:    "",
		UnzipOfflinePackage:   unzipPkgCmd,
		CopyKubernentesBinary: copyKubeBinary,
		LoadDockerImages:      loadDockerImages,
		KubeEnvironmentInit:   "",
		JoinWorkerEtcHostsSet: echoWorkerHostCmd,
		WorkerStaticPodCreate: "",
		KubeadmJoinWorker:     "",
	}

}
