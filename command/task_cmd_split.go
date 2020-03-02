package command

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// inspect command
func TaskInspectCommandSplit(cobraCmd, taskName, taskCommand string, kubeInstaller *KubeInstaller, cobraFlag FlagCobra) {
	switch taskName {
	// ping each other based on host IP to prove host network interoperability
	case "check_system_network":
		var newCmd string
		for _, v := range kubeInstaller.Hosts {
			newCmd = newCmd + fmt.Sprintf("ping %s -c 3;", v)
		}
		taskCommand = newCmd
		ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.Hosts, taskCommand, cobraFlag)
	default:
		ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.Hosts, taskCommand, cobraFlag)
	}
}

// bootstrap command
func TaskBootstrapCommandSplit(cobraCmd, taskName, taskCommand string, kubeInstaller *KubeInstaller,
	cobraFlag FlagCobra) {
	switch taskName {
	case "bootstrap_chrony_server":
		// determine whether the time synchronization server is equal to the default value. If not,
		// follow the specified configuration.
		if cobraFlag.ChronyServer != "pool.ntp.org" {
			chronyCfg := fmt.Sprintf("server %s iburst", cobraFlag.ChronyServer)
			chronyServer := fmt.Sprintf(`grep '%s'  /etc/chrony.conf || sed -i 's/^.*centos.pool.ntp.org/#&/g' /etc/chrony.conf && sed -i '/#server 3.centos.pool.ntp.org iburst/a\\%s' /etc/chrony.conf`, chronyCfg, chronyCfg)
			taskCommand = chronyServer
		}
		ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.Hosts, taskCommand, cobraFlag)
	default:
		ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.Hosts, taskCommand, cobraFlag)
	}
}

// init command
func TaskInitCommandSplit(cobraCmd, taskName, taskCommand string, kubeInstaller *KubeInstaller, cobraFlag FlagCobra) {
	switch taskName {
	case "send_offline_package":
		ExecSFTPCommandMain("send_offline_package", kubeInstaller.Hosts, cobraFlag)
	case "kube_environment_init":
		kubeInstaller.kubeEnvironmentInit(kubeInstaller.Hosts, cobraCmd, cobraFlag)
	case "kubeadm_init_config_gen":
		// kubeInstaller first host
		taskCommand = kubeInstaller.kubeadmConfigGenerate(cobraFlag)
		SingleHostTaskCommandExec(taskName, kubeInstaller.Master[0], taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(
			taskName)
	case "configure_master0_etc_hosts":
		SingleHostTaskCommandExec(taskName, kubeInstaller.Master[0], taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	case "kubeadm_init_master0":
		SingleHostTaskCommandExec(taskName, kubeInstaller.Master[0], taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	case "copy_kubeadm_master0_config":
		fmt.Println(taskCommand)
		SingleHostTaskCommandExec(taskName, kubeInstaller.Master[0], taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	case "network_plugin_config_gen":
		// kubeInstaller first host
		_, taskCommand = kubeInstaller.networkPluginConfigGen(cobraFlag)
		SingleHostTaskCommandExec(taskName, kubeInstaller.Master[0], taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	case "network_plugin_install":
		taskCommand = kubeInstaller.networkPluginInstall(cobraFlag)
		SingleHostTaskCommandExec(taskName, kubeInstaller.Master[0], taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	case "join_master_etc_hosts_set":
		if len(kubeInstaller.Master) > 1 {
			ExecSSHCommandMain(taskName, kubeInstaller.Master[1:], taskCommand,
				cobraFlag).ResultOutputCheckAndProcessExit(taskName)
		}
	case "kubeadm_join_master":
		if len(kubeInstaller.Master) > 1 {
			// kubeadm get join token，tokenCaCertHash, certificateKey params from master0
			joinToken, tokenCaCertHash, certificateKey := kubeInstaller.getJoinParamFromMaster0(cobraFlag)
			taskCommand = kubeInstaller.joinMasterCmd(joinToken, tokenCaCertHash, certificateKey, cobraFlag)
			ExecSSHCommandMain(taskName, kubeInstaller.Master[1:], taskCommand,
				cobraFlag).ResultOutputCheckAndProcessExit(taskName)
		}
	case "sed_join_master_etc_hosts":
		if len(kubeInstaller.Master) > 1 {
			for _, master := range kubeInstaller.Master[1:] {
				taskCommand = fmt.Sprintf(`sed "s/%s/%s/g" -i /etc/hosts`, kubeInstaller.Master[0], master)
				SingleHostTaskCommandExec(taskName, master, taskCommand,
					cobraFlag).ResultOutputCheckAndProcessExit(taskName)
			}
		}
	case "copy_kubernentes_config":
		if len(kubeInstaller.Master) > 1 {
			ExecSSHCommandMain(taskName, kubeInstaller.Master[1:], taskCommand,
				cobraFlag).ResultOutputCheckAndProcessExit(taskName)
		}
	case "join_worker_etc_hosts_set":
		ExecSSHCommandMain(taskName, kubeInstaller.Worker, taskCommand,
			cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	case "worker_static_pod_create":
		kubeInstaller.lvScareStaticPodYamlGen(kubeInstaller.Worker, cobraFlag)
	case "kubeadm_join_worker":
		// kubeadm get join token，tokenCaCertHash, certificateKey params from master0
		joinToken, tokenCaCertHash, certificateKey := kubeInstaller.getJoinParamFromMaster0(cobraFlag)
		taskCommand = kubeInstaller.joinWorkerCmd(joinToken, tokenCaCertHash, certificateKey, cobraFlag)
		ExecSSHCommandMain(taskName, kubeInstaller.Worker, taskCommand,
			cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	default:
		ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.Hosts, taskCommand, cobraFlag)
	}

}

// reset command
func TaskResetCommandSplit(cobraCmd, taskName, taskCommand string, kubeInstaller *KubeInstaller, cobraFlag FlagCobra) {
	switch taskName {
	case "clean_kube_all":
		if cobraFlag.RemoveAll {
			CleanKubeAll(cobraCmd, kubeInstaller, cobraFlag)
		}
	case "clean_kube_component":
		if cobraFlag.RemoveKubeComponents {
			ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.Hosts, taskCommand, cobraFlag)
		}
	case "clean_docker_images":
		if cobraFlag.RemoveContainerImages {
			ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.Hosts, taskCommand, cobraFlag)
		}
	case "clean_install_pkg":
		if cobraFlag.RemoveInstallPkg {
			if cobraFlag.PkgPath != "" {
				baseDir := filepath.Dir(cobraFlag.PkgPath)
				taskCommand = fmt.Sprintf("rm -rf %s/kube && rm -rf %s", baseDir, cobraFlag.PkgPath)
				ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.Hosts, taskCommand, cobraFlag)
			}
			log.Println("Program tidbkube clean install pkg need set flag --pkg-path (install offline path)")
			os.Exit(1)
		}
	default:
		ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.Hosts, taskCommand, cobraFlag)
	}
}

// join master command
func TaskJoinMasterCommandSplit(cobraCmd, taskName, taskCommand string, kubeInstaller *KubeInstaller, cobraFlag FlagCobra) {
	switch taskName {
	case "send_offline_package":
		ExecSFTPCommandMain("send_offline_package", kubeInstaller.JoinNode, cobraFlag)
	case "kube_environment_init":
		kubeInstaller.kubeEnvironmentInit(kubeInstaller.JoinNode, cobraCmd, cobraFlag)
	case "join_master_etc_hosts_set":
		ExecSSHCommandMain(taskName, kubeInstaller.JoinNode, taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	case "kubeadm_join_master":
		// kubeadm get join token，tokenCaCertHash, certificateKey params from master0
		joinToken, tokenCaCertHash, certificateKey := kubeInstaller.getJoinParamFromMaster0(cobraFlag)
		taskCommand = kubeInstaller.joinMasterCmd(joinToken, tokenCaCertHash, certificateKey, cobraFlag)
		ExecSSHCommandMain(taskName, kubeInstaller.JoinNode, taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(
			taskName)
	case "sed_join_master_etc_hosts":
		for _, master := range kubeInstaller.JoinNode {
			taskCommand = fmt.Sprintf(`sed "s/%s/%s/g" -i /etc/hosts`, kubeInstaller.Master[0], master)
			SingleHostTaskCommandExec(taskName, master, taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
		}
	case "copy_kubernentes_config":
		ExecSSHCommandMain(taskName, kubeInstaller.JoinNode, taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	case "add_worker_node_ipvs_rule":
		kubeInstaller.addWorkerNodeIPVSRule(cobraFlag)
	case "update_worker_ipvs_static_pod":
		kubeInstaller.updateIPVSStaticPodYamlFile(cobraFlag)
	default:
		ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.JoinNode, taskCommand, cobraFlag)
	}
}

// join worker command
func TaskJoinWorkerCommandSplit(cobraCmd, taskName, taskCommand string, kubeInstaller *KubeInstaller, cobraFlag FlagCobra) {
	switch taskName {
	case "send_offline_package":
		ExecSFTPCommandMain("send_offline_package", kubeInstaller.JoinNode, cobraFlag)
	case "kube_environment_init":
		kubeInstaller.kubeEnvironmentInit(kubeInstaller.JoinNode, cobraCmd, cobraFlag)
	case "join_worker_etc_hosts_set":
		ExecSSHCommandMain(taskName, kubeInstaller.JoinNode, taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	case "worker_static_pod_create":
		kubeInstaller.lvScareStaticPodYamlGen(kubeInstaller.JoinNode, cobraFlag)
	case "kubeadm_join_worker":
		// kubeadm get join token，tokenCaCertHash, certificateKey params from master0
		joinToken, tokenCaCertHash, certificateKey := kubeInstaller.getJoinParamFromMaster0(cobraFlag)
		taskCommand = kubeInstaller.joinWorkerCmd(joinToken, tokenCaCertHash, certificateKey, cobraFlag)
		ExecSSHCommandMain(taskName, kubeInstaller.JoinNode, taskCommand, cobraFlag).ResultOutputCheckAndProcessExit(taskName)
	default:
		ExecTaskCommandMain(cobraCmd, taskName, kubeInstaller.JoinNode, taskCommand, cobraFlag)
	}
}
