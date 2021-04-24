package command

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wentaojin/tidbkube/network"

	"github.com/wentaojin/tidbkube/util"
)

// kubeadm command type
type CommandType string

const (
	InitMaster CommandType = "initMaster"
	JoinMaster CommandType = "joinMaster"
	JoinWorker CommandType = "joinWorker"
)

// kubeadm k8s installer
type KubeInstaller struct {
	Hosts    []string
	Master   []string
	Worker   []string
	JoinNode []string
}

// Parse the tidbkube flag master and worker host IP values, used for the operation of a series of commands of the program
func TidbKubeInstaller(cobraFlag FlagCobra) *KubeInstaller {
	kubeInstanller := hostIPRepeatInspect(cobraFlag)
	return kubeInstanller
}

func (k *KubeInstaller) kubeEnvironmentInit(hosts []string, cobraCmd string, cobraFlag FlagCobra) {
	if len(hosts) == 0 {
		log.Println("Func kubeEnvironmentInit:::Kubernentes environment init hosts slice Can't Null.")
		os.Exit(1)
	}
	for _, host := range hosts {
		setKubeletSystemService := fmt.Sprintf(`cat <<EOF > /etc/systemd/system/kubelet.service
%v
EOF`, KubeletServiceTemp)
		SingleHostTaskCommandExec("set_kubelet_system_service", host, setKubeletSystemService, cobraFlag).ResultOutputCheckAndProcessExit("set_kubelet_system_service")

		mkdirKubeletServiceDir := fmt.Sprintf("if [ ! -d /etc/systemd/system/kubelet.service.d ]; then mkdir -p /etc/systemd/system/kubelet.service.d; fi")
		SingleHostTaskCommandExec("mkdir_kubelet_service_dir", host, mkdirKubeletServiceDir, cobraFlag).ResultOutputCheckAndProcessExit("mkdir_kubelet_service_dir")

		copyKubeadmConfigFile := fmt.Sprintf(`cat <<EOF > /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
%v
EOF`, Kubeadm10ConfTemp)
		SingleHostTaskCommandExec("copy_kubeadm_config_file", host, copyKubeadmConfigFile, cobraFlag).ResultOutputCheckAndProcessExit("copy_kubeadm_config_file")

		mkdirKubeletDir := fmt.Sprintf("if [ ! -d /var/lib/kubelet ]; then mkdir -p /var/lib/kubelet; fi")
		SingleHostTaskCommandExec("mkdir_kubelet_dir", host, mkdirKubeletDir, cobraFlag).ResultOutputCheckAndProcessExit("mkdir_kubelet_dir")

		getCgroupDriver := fmt.Sprintf("systemctl restart docker && echo $(docker info|grep Cg)|awk -F ':' '{print $2}'|awk '{print $1}'")
		sshResult := SingleHostTaskCommandExec("get_docker_cgroup_driver", host, getCgroupDriver, cobraFlag)
		sshResult.ResultOutputCheckAndProcessExit("get_docker_cgroup_driver")
		cgroupDriver := sshResult.ResultOutputCgroupDriver()

		generateKubeletConfig := fmt.Sprintf(`cat <<EOF > /var/lib/kubelet/config.yaml
%v
EOF`, string(TemplateKubelet(cgroupDriver)))
		SingleHostTaskCommandExec("generate_kubelet_config", host, generateKubeletConfig, cobraFlag).ResultOutputCheckAndProcessExit("generate_kubelet_config")

		generateSysconfigKubelet := fmt.Sprintf(`cat <<EOF > /etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--cgroup-driver=%v
EOF`, cgroupDriver)
		SingleHostTaskCommandExec("generate_kubelet_extra_args", host, generateSysconfigKubelet, cobraFlag).ResultOutputCheckAndProcessExit("generate_kubelet_extra_args")

		systemDaemonReload := fmt.Sprintf("systemctl daemon-reload")
		SingleHostTaskCommandExec("system_daemon_reload", host, systemDaemonReload, cobraFlag).ResultOutputCheckAndProcessExit("system_daemon_reload")

		systemEnableKubelet := fmt.Sprintf("systemctl enable kubelet")
		SingleHostTaskCommandExec("system_enable_kubelet", host, systemEnableKubelet, cobraFlag).ResultOutputCheckAndProcessExit("system_enable_kubelet")
	}
}

func (k *KubeInstaller) kubeadmConfigGenerate(cobraFlag FlagCobra) (taskCommand string) {
	var templateData string
	if cobraFlag.KubeadmConfigFile == "" {
		templateData = string(Template(cobraFlag))
		PrintlnKubeadmConfig(cobraFlag)
	} else {
		fileData, err := ioutil.ReadFile(cobraFlag.KubeadmConfigFile)
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Func KubeadmConfigGenerate:::Flag kubeadm-config file read failed: %v\n", err)
				os.Exit(1)
			}
		}()
		if err != nil {
			panic(err)
		}
		templateData = string(fileData)
		log.Printf("Generate kubeadm config file,As follow:\n%s", templateData)
	}

	fileDirName := filepath.Dir(cobraFlag.PkgPath)
	//taskCommand := fmt.Sprintf("echo \""+templateData+"\" > %s/kubeadm-config.yaml", fileDirName)
	taskCommand = fmt.Sprintf(`cat << EOF > %s/kubeadm-config.yaml
%v
EOF`, fileDirName, templateData)

	return
}

func (k *KubeInstaller) kubeMaster0EtcHostSet(cobraFlag FlagCobra) (echoHostCmd string) {
	echoHostCmd = fmt.Sprintf("echo %s %s >> /etc/hosts", k.Master[0], cobraFlag.ApiServer)
	return
}

func (k *KubeInstaller) kubeMaster0Init(cobraFlag FlagCobra, joinToken, tokenCaCertHash,
	certificateKey string) (initMasterCmd string) {
	initMasterCmd = k.kubeadmCommand(cobraFlag, InitMaster, joinToken, tokenCaCertHash, certificateKey)
	return
}

func (k *KubeInstaller) kubeMaster0DirCreate(cobraFlag FlagCobra) (kubeadmMaster0ConfigCmd string) {
	// sudo user todo
	//kubeadmMaster0ConfigCmd = `sudo mkdir -p $HOME/.kube && sudo /bin/cp -rf /etc/kubernetes/admin.conf $HOME/.
	// kube/config && sudo chown $(id -u):$(id -g) $HOME/.kube/config`
	kubeadmMaster0ConfigCmd = `mkdir -p $HOME/.kube && /bin/cp -rf /etc/kubernetes/admin.conf $HOME/.kube/config`
	return
}

func (k *KubeInstaller) networkPluginConfigGen(cobraFlag FlagCobra) (netPluginFile, netPluginGenCmd string) {
	if cobraFlag.WithoutCNI {
		if cobraFlag.NetworkPluginFile != "" {
			fileData, err := ioutil.ReadFile(cobraFlag.NetworkPluginFile)
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Func NetworkPluginConfigGen:::Flag plugin-config file read failed: %v\n", err)
					os.Exit(1)
				}
			}()
			if err != nil {
				panic(err)
			}
			templateData := string(fileData)
			fileDirName := filepath.Dir(cobraFlag.PkgPath)
			fileName := fmt.Sprintf("%s/network-plugin-config.yaml", fileDirName)
			netPluginGenCmd = fmt.Sprintf(`cat << EOF > %s
%v
EOF`, fileName, templateData)

			return "yes", netPluginGenCmd
		}
		return "no", netPluginGenCmd
	}

	netYaml := network.NewNetwork(cobraFlag.NetworkPlugin, network.MetaData{Interface: cobraFlag.NetworkInterface,
		PodCIDR: cobraFlag.PodCIDR}).Manifests("")
	fileDirName := filepath.Dir(cobraFlag.PkgPath)
	fileName := fmt.Sprintf("%s/network-plugin-config.yaml", fileDirName)
	netPluginGenCmd = fmt.Sprintf(`cat << EOF > %s
%v
EOF`, fileName, netYaml)
	return "yes", netPluginGenCmd
}

func (k *KubeInstaller) networkPluginInstall(cobraFlag FlagCobra) (netPluginInstallCmd string) {
	netPluginFile, _ := k.networkPluginConfigGen(cobraFlag)
	switch netPluginFile {
	case "no":
		log.Println(`Func NetworkPluginInstall:::Cobra command init flag --without-cni is true,
so we not install calico , install it by yourself`)
		log.Println(`Func NetworkPluginInstall:::Cobra command init flag --net-plugin-config and --net-plugin-name
Need configure, Please reset kubernentes environment and reinstall`)
		os.Exit(1)
	default:
		//cmd = `kubectl apply -f /root/kube/conf/net/calico.yaml || true`
		fileDirName := filepath.Dir(cobraFlag.PkgPath)
		fileName := fmt.Sprintf("%s/network-plugin-config.yaml", fileDirName)
		netPluginInstallCmd = fmt.Sprintf(`kubectl apply -f %s || true`, fileName)
	}
	return
}

func (k *KubeInstaller) joinMasterEtcHost(cobraFlag FlagCobra) (joinMasterEtcHostCmd string) {
	joinMasterEtcHostCmd = fmt.Sprintf("echo %s %s >> /etc/hosts", k.Master[0], cobraFlag.ApiServer)
	return
}

func (k *KubeInstaller) joinMasterCmd(joinToken, tokenCaCertHash, certificateKey string,
	cobraFlag FlagCobra) (joinMasterCmd string) {
	joinMasterCmd = k.kubeadmCommand(cobraFlag, JoinMaster, joinToken, tokenCaCertHash, certificateKey)
	return
}

func (k *KubeInstaller) joinMasterConfig(cobraFlag FlagCobra) (copyK8sConfCmd string) {
	// sudo user todo
	//copyK8sConfCmd = `sudo mkdir -p $HOME/.kube && sudo cp -i /etc/kubernetes/admin.conf $HOME/.
	// kube/config && sudo chown $(id -u):$(id -g) $HOME/.kube/config`
	copyK8sConfCmd = `mkdir -p $HOME/.kube && /bin/cp -rf /etc/kubernetes/admin.conf $HOME/.kube/config`
	//pkgDir := filepath.Dir(cobraFlag.PkgPath)
	//cleanInstall := fmt.Sprintf("rm -rf %s", pkgDir)
	//ExecSSHCommandMain("clean_install_package_path", k.Master[1:], cleanInstall,
	//	cobraFlag).ResultOutputCheckAndProcessExit("clean_install_package_path")
	return
}

func (k *KubeInstaller) joinWorkerEtcHost(cobraFlag FlagCobra) (echoWorkerHostCmd string) {
	echoWorkerHostCmd = fmt.Sprintf("echo %s %s >> /etc/hosts", cobraFlag.VirtualIP, cobraFlag.ApiServer)
	return
}

func (k *KubeInstaller) lvScareStaticPodYamlGen(worker []string, cobraFlag FlagCobra) {
	// ipvs rule create and static pod yaml generate
	lvScareCmd := k.joinWorkerIpvsRuleCreate(cobraFlag)
	ExecSSHCommandMain("ipvs_rule_create", worker, lvScareCmd, cobraFlag).ResultOutputCheckAndProcessExit(
		"ipvs_rule_create")

	// ipvs static pod yaml file generate
	mkdirStaticPodDir := fmt.Sprintf("if [ ! -d /etc/kubernetes/manifests ]; then mkdir -p /etc/kubernetes/manifests; fi")
	ExecSSHCommandMain("mkdir_node_static_pod_dir", worker, mkdirStaticPodDir,
		cobraFlag).ResultOutputCheckAndProcessExit("mkdir_node_static_pod_dir")

	// println ipvs static pod yaml file
	PrintlnIPVSStaticPodConfig(cobraFlag)

	// generate ipvs static pod yaml file to host /etc/kubernetes/manifests/ dir
	genStaticPodYaml := fmt.Sprintf(`cat > /etc/kubernetes/manifests/kube-ipvs-static.yaml  <<EOF
%v
EOF`, string(TemplateIPVS(cobraFlag)))
	ExecSSHCommandMain("gen_node_static_pod_yaml", worker, genStaticPodYaml,
		cobraFlag).ResultOutputCheckAndProcessExit("gen_node_static_pod_yaml")
}

func (k *KubeInstaller) joinWorkerIpvsRuleCreate(cobraFlag FlagCobra) (lvScareCmd string) {
	// LVScare A lightweight LVS baby care, support ipvs health check
	// If ipvs real server is unavilible, remove it, if real server return to normal, add it back. This is useful for kubernetes master HA.
	//If it is not the control node and is not a single master, then create an ipvs rule bofore join worker node.
	// The control node does not need to be created, even its own apiserver can be used.
	// Then kubeadm join can use VIP:6443 instead real masters.
	// Finally Run lvscare as a static pod on every kubernetes worker node.

	//vip := fmt.Sprintf("%s:6443", cobraFlag.VirtualIP)
	//var masters []string
	//for _, m := range k.Master {
	//	masterIP := fmt.Sprintf("%s:6443", m)
	//	masters = append(masters, masterIP)
	//}
	//err := create.VsAndRsCreate(vip, masters)
	//if err != nil {
	//	log.Fatalf("Func JoinWorkers:::LVScare create ipvs rule falied: %v.\n", err.Error())
	//	os.Exit(1)
	//}
	vip := fmt.Sprintf("%s:6443", cobraFlag.VirtualIP)
	var mr string
	for _, m := range k.Master {
		masterIP := fmt.Sprintf(" --rs %s:6443", m)
		mr += masterIP
	}
	// kubernents worker node ipvs rule create
	lvScareCmd = fmt.Sprintf("lvscare create --vs %s %s", vip, mr)
	return
}

func (k *KubeInstaller) joinWorkerCmd(joinToken, tokenCaCertHash, certificateKey string,
	cobraFlag FlagCobra) (joinWorkerCmd string) {
	joinWorkerCmd = k.kubeadmCommand(cobraFlag, JoinWorker, joinToken, tokenCaCertHash, certificateKey)
	//pkgDir := filepath.Dir(cobraFlag.PkgPath)
	//cleanInstall := fmt.Sprintf("rm -rf %s", pkgDir)
	//ExecSSHCommandMain("clean_install_package_path", k.Worker, cleanInstall,
	//	cobraFlag).ResultOutputCheckAndProcessExit("clean_install_package_path")
	return
}

func (k *KubeInstaller) getJoinParamFromMaster0(cobraFlag FlagCobra) (joinToken, tokenCaCertHash,
	certificateKey string) {
	// kubeadm get join token，tokenCaCertHash, certificateKey params from master0
	joinToken, tokenCaCertHash = k.generatorToken(cobraFlag)
	certificateKey = k.generatorCerts(cobraFlag)
	return
}

func (k *KubeInstaller) kubeadmCommand(cobraFlag FlagCobra, commandType CommandType, joinToken, tokenCaCertHash,
	certificateKey string) (cmd string) {
	fileDirName := filepath.Dir(cobraFlag.PkgPath)
	cmds := make(map[CommandType]string)
	cmds = map[CommandType]string{
		InitMaster: fmt.Sprintf(`kubeadm init --config=%s/kubeadm-config.yaml --experimental-upload-certs`,
			fileDirName),
		JoinMaster: fmt.Sprintf(`kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --experimental-control-plane --certificate-key %s`, k.Master[0], joinToken, tokenCaCertHash, certificateKey),
		JoinWorker: fmt.Sprintf(`kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --ignore-preflight-errors=DirAvailable--etc-kubernetes-manifests`,
			cobraFlag.VirtualIP,
			joinToken,
			tokenCaCertHash),
	}
	// other kubernentes version
	//todo
	if util.K8sVersionConvertToInt(cobraFlag.K8sVersion) >= 115 {
		cmds[InitMaster] = fmt.Sprintf(`kubeadm init --config=%s/kubeadm-config.yaml --upload-certs`, fileDirName)
		cmds[JoinMaster] = fmt.Sprintf(`kubeadm join %s:6443 --token %s --discovery-token-ca-cert-hash %s --control-plane --certificate-key %s`, k.Master[0], joinToken, tokenCaCertHash, certificateKey)
	}

	v, ok := cmds[commandType]
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Func KubeadmCommand:::Fetch kubeadm command error.")
		}
	}()
	if !ok {
		panic(1)
	}
	return v
}

// GeneratorCerts function
func (k *KubeInstaller) generatorCerts(cobraFlag FlagCobra) (certificateKey string) {
	getParamsCmd := fmt.Sprintf(`kubeadm init phase upload-certs --upload-certs`)
	taskName := "kubeadm_get_join_master_cert_from_master0"
	sshResultLog := SingleHostTaskCommandExec(taskName, k.Master[0], getParamsCmd, cobraFlag)
	sshResultLog.ResultOutputCheckAndProcessExit(taskName)
	certificateKey = sshResultLog.ResultOutputKubeadmJoinMasterCert()
	return
}

// GeneratorToken function
func (k *KubeInstaller) generatorToken(cobraFlag FlagCobra) (joinToken, tokenCaCertHash string) {
	getParamsCmd := fmt.Sprintf("kubeadm token create --print-join-command")
	taskName := "kubeadm_get_join_params_from_master0"
	sshResultLog := SingleHostTaskCommandExec(taskName, k.Master[0], getParamsCmd, cobraFlag)
	sshResultLog.ResultOutputCheckAndProcessExit(taskName)
	joinToken, tokenCaCertHash = sshResultLog.ResultOutputKubeadmJoinParams()
	return
}

func (k *KubeInstaller) addWorkerNodeIPVSRule(cobraFlag FlagCobra) {
	// add new master node to all normal exist worker node
	for _, worker := range k.Worker {
		ipvsadmCmd := fmt.Sprintf(`ipvsadm -S`)
		SingleHostTaskCommandExec("view_worker_node_ipvs_add_rule", worker, ipvsadmCmd,
			cobraFlag).ResultOutputCheckAndProcessExit("view_worker_node_ipvs_add_rule")
		for _, joinMaster := range k.JoinNode {
			addIpvsadmCmd := fmt.Sprintf(`ipvsadm -a -t %s:6443 -r %s:6443  -m -w 1`, cobraFlag.VirtualIP, joinMaster)
			SingleHostTaskCommandExec("add_worker_node_ipvs_rule", worker, addIpvsadmCmd,
				cobraFlag).ResultOutputCheckAndProcessExit("add_worker_node_ipvs_rule")
		}
	}
}

func (k *KubeInstaller) updateIPVSStaticPodYamlFile(cobraFlag FlagCobra) {
	// update all normal exist worker node ipvs static pod yaml rule,
	// add new master real server
	for _, worker := range k.Worker {
		for _, joinMaster := range k.JoinNode {
			cmd := fmt.Sprintf(`sed '/- https/a\    - %s:6443' -i /etc/kubernetes/manifests/kube-ipvs-static.yaml`, joinMaster)
			SingleHostTaskCommandExec("add_master_static_pod_yaml", worker, cmd,
				cobraFlag).ResultOutputCheckAndProcessExit("add_master_static_pod_yaml")

			cmd = `sed '/- https/a\    - --rs' -i /etc/kubernetes/manifests/kube-ipvs-static.yaml`
			SingleHostTaskCommandExec("update_static_pod_yaml", worker, cmd, cobraFlag).ResultOutputCheckAndProcessExit(
				"update_static_pod_yaml")
		}
	}
}

// hostIPRepeatInspect function,cobra flag master、worker value repeat inspect
func hostIPRepeatInspect(cobraFlag FlagCobra) *KubeInstaller {
	// all master 节点
	masters := util.ParseIPSegment(cobraFlag.MasterIP)

	if len(masters) == 0 {
		log.Fatalf("Func hostIPRepeatInspect:::K8s Master host IP not allow empty,Please view command help (--help)")
		os.Exit(1)
	}

	// all worker 节点
	workers := util.ParseIPSegment(cobraFlag.WorkerIP)

	if len(workers) == 0 {
		log.Fatalf("Func hostIPRepeatInspect:::K8s Worker host IP not allow empty,Please view command help (--help)")
		os.Exit(1)
	}

	// determine if the program master and worker value IP conflict
	status, allRepeatValue, _, _ := util.StringSliceCountValues(masters)
	if !status {
		log.Printf("Func hostIPRepeatInspect:::Host system master host IP exist repeat value null,Require host IP non-null")
		os.Exit(1)
	}
	if len(allRepeatValue) != 0 {
		log.Printf("Func hostIPRepeatInspect:::Host system master host IP exist repeat value %v,"+
			"Require host IP non-repeat", allRepeatValue)
		os.Exit(1)
	}

	status, allRepeatValue, _, _ = util.StringSliceCountValues(workers)
	if !status {
		log.Printf("Func hostIPRepeatInspect:::Host system master host IP exist repeat value null,Require host IP non-null")
		os.Exit(1)
	}

	if len(allRepeatValue) != 0 {
		log.Printf("Func hostIPRepeatInspect:::Host system master host IP exist repeat value %v,"+
			"Require host IP non-repeat", allRepeatValue)
		os.Exit(1)
	}

	allHosts := append(masters, workers...)

	status, allRepeatValue, _, _ = util.StringSliceCountValues(allHosts)
	if !status {
		log.Printf("Func hostIPRepeatInspect:::Host system master and worker host IP exist repeat value null," +
			"Require host IP non-null")
		os.Exit(1)
	}

	if len(allRepeatValue) != 0 {
		log.Printf("Func hostIPRepeatInspect:::Host system master and worker host IP exist repeat value %v,"+
			"Require host IP non-repeat", allRepeatValue)
		os.Exit(1)
	}

	if len(allHosts) == 0 {
		log.Fatalf("Func hostIPRepeatInspect:::Host not allow empty")
		os.Exit(1)
	}

	// get join master and worker node host IP
	joinNode := util.ParseIPSegment(cobraFlag.JoinNodeIP)

	installer := &KubeInstaller{
		Hosts:    allHosts,
		Master:   masters,
		Worker:   workers,
		JoinNode: joinNode,
	}
	return installer
}
