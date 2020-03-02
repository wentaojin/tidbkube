package command

import (
	"fmt"
	"log"
	"path/filepath"
)

// Init kubernentes cluster
func InitKubernentesCluster(cobraFlag FlagCobra) {
	cobraFlag.PrintFlagCobraConfig()
	kubeInstaller := TidbKubeInstaller(cobraFlag)
	taskInitCommand := TaskInitCommand(kubeInstaller, cobraFlag)
	initCommandTask := &InitCommandTask{}
	if cobraFlag.ListTask {
		flagListTaskCommandExec(initCommandTask)
	} else {
		log.Printf("Program starts host-level kubernentes install, Please wait a moment.\n")
		TaskCommandExec("init", cobraFlag, taskInitCommand)
		log.Println("Program tidbkube command init exec success.")
	}
}

// InitCommandTask struct show machine ssh shell init command,mainly used for kubernentes env init
type InitCommandTask struct {
	// offline package init
	SendOfflinePackage    string `json:"send_offline_package"`
	UnzipOfflinePackage   string `json:"unzip_offline_package"`
	CopyKubernentesBinary string `json:"copy_kubernentes_binary"`
	LoadDockerImages      string `json:"load_docker_images"`
	// kubernentes env init
	KubeEnvironmentInit string `json:"kube_environment_init"`
	// kubernentes install
	KubeadmInitConfigGen     string `json:"kubeadm_init_config_gen"`
	ConfigureMaster0EtcHosts string `json:"configure_master0_etc_hosts"`
	KubeadmInitMaster0       string `json:"kubeadm_init_master0"`
	CopyKubeadmMaster0Config string `json:"copy_kubeadm_master0_config"`
	NetworkPluginConfigGen   string `json:"network_plugin_config_gen"`
	NetworkPluginInstall     string `json:"network_plugin_install"`
	JoinMasterEtcHostsSet    string `json:"join_master_etc_hosts_set"`
	KubeadmJoinMaster        string `json:"kubeadm_join_master"`
	SedJoinMasterEtcHosts    string `json:"sed_join_master_etc_hosts"`
	CopyKubernentesConfig    string `json:"copy_kubernentes_config"`
	JoinWorkerEtcHostsSet    string `json:"join_worker_etc_hosts_set"`
	WorkerStaticPodCreate    string `json:"worker_static_pod_create"`
	KubeadmJoinWorker        string `json:"kubeadm_join_worker"`
}

// TaskInitCommand function
func TaskInitCommand(kubeInstaller *KubeInstaller, cobraFlag FlagCobra) *InitCommandTask {
	baseDir := filepath.Dir(cobraFlag.PkgPath)
	fileName := filepath.Base(cobraFlag.PkgPath)
	unzipPkgCmd := fmt.Sprintf("cd %s && tar zxvf %s", baseDir, fileName)
	copyKubeBinary := fmt.Sprintf("cp  %s/kube/bin/* /usr/bin", baseDir)
	loadDockerImages := fmt.Sprintf("docker load -i %s/kube/images/images.tar", baseDir)

	configureMaster0EtcHosts := kubeInstaller.kubeMaster0EtcHostSet(cobraFlag)
	// kubeadm master0 init not need joinToken, tokenCaCertHash, certificateKey params
	kubeadmInitMaster0 := kubeInstaller.kubeMaster0Init(cobraFlag, "", "", "")
	CopyKubeadmMaster0Config := kubeInstaller.kubeMaster0DirCreate(cobraFlag)

	joinMasterEtcHostCmd := kubeInstaller.joinMasterEtcHost(cobraFlag)
	copyK8sConfCmd := kubeInstaller.joinMasterConfig(cobraFlag)
	echoWorkerHostCmd := kubeInstaller.joinWorkerEtcHost(cobraFlag)

	return &InitCommandTask{
		SendOfflinePackage:       "",
		UnzipOfflinePackage:      unzipPkgCmd,
		CopyKubernentesBinary:    copyKubeBinary,
		LoadDockerImages:         loadDockerImages,
		KubeEnvironmentInit:      "",
		KubeadmInitConfigGen:     "",
		ConfigureMaster0EtcHosts: configureMaster0EtcHosts,
		KubeadmInitMaster0:       kubeadmInitMaster0,
		CopyKubeadmMaster0Config: CopyKubeadmMaster0Config,
		NetworkPluginConfigGen:   "",
		NetworkPluginInstall:     "",
		JoinMasterEtcHostsSet:    joinMasterEtcHostCmd,
		KubeadmJoinMaster:        "",
		SedJoinMasterEtcHosts:    "",
		CopyKubernentesConfig:    copyK8sConfCmd,
		JoinWorkerEtcHostsSet:    echoWorkerHostCmd,
		WorkerStaticPodCreate:    "",
		KubeadmJoinWorker:        "",
	}

}
