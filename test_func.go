package main

import (
	"fmt"

	"github.com/WentaoJin/tidbkube/command"
)

func main() {

	var cobraFlag command.FlagCobra

	cobraFlag.SSHUser = "root"
	cobraFlag.SSHPort = "22"
	cobraFlag.SSHPassword = "pingcap!@#"

	copyKubeadmConfigFile := fmt.Sprintf(`cat <<EOF > /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
%v
EOF`, command.Kubeadm10ConfTemp)
	command.SingleHostTaskCommandExec("copy_kubeadm_config_file", "172.16.5.89", copyKubeadmConfigFile,
		cobraFlag).ResultOutputCheckAndProcessExit(
		"copy_kubeadm_config_file")
	//command.PrintlnKubeadmConfig()
	//host := "172.16.5.83"
	//port := "22"
	//user := "root"
	//password := "pingcap!@#"
	//locapath := "C:\\Marvin\\Projects\\goModules\\tidbkube\\command\\cmd_bootstrap.go"
	//destpath := "/home/tidb/marvin/cmd_bootstrap.go"
	//util.SFTPCopy(host, port, user, password, "", "", locapath, destpath)
}
