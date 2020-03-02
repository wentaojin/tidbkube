package cli

import (
	"fmt"

	"github.com/WentaoJin/tidbkube/command"
	"github.com/spf13/cobra"
)

// kubeCmd define cli program command
var kubeCmd = &cobra.Command{
	Use:     "kube",
	Aliases: []string{"kb"},
	Short:   `The command kube is used for initialize kubernentes cluster installation and join adding kubernentes nodes`,
	Long: fmt.Sprintf("%v\nThe command kube is used for kubernentes init to initialize the cluster installation and"+
		" join, and add kubernentes nodes", headerStyle),
	PreRun: func(cmd *cobra.Command, args []string) {
		initProgramRun()
	},
}

func init() {
	rootCmd.AddCommand(kubeCmd)
	kubeCmd.PersistentFlags().StringVarP(&command.CobraFlag.KubeadmConfigFile, "kubeadm-config", "c", "",
		"Kubeadm-config.yaml file")
	kubeCmd.PersistentFlags().StringVar(&command.CobraFlag.ApiServer, "apiserver", "apiserver.cluster.local",
		"Specify a DNS name for the control plane")
	kubeCmd.PersistentFlags().StringVarP(&command.CobraFlag.NetworkPlugin, "net-plugin-name", "n", "calico", "Cni plugin, calico..")
	kubeCmd.PersistentFlags().StringVar(&command.CobraFlag.NetworkInterface, "net-interface", "eth.*|em.*",
		"Specify a network interface name of IP address")
	kubeCmd.PersistentFlags().StringVar(&command.CobraFlag.NetworkPluginFile, "net-plugin-config", "",
		"Kubernentes network plugin yaml config file (for example: /root/kube-flannel.yaml)")
	kubeCmd.PersistentFlags().StringVar(&command.CobraFlag.VirtualIP, "vip", "10.103.97.2", "kubernentes virtual ip")
	kubeCmd.PersistentFlags().StringVar(&command.CobraFlag.PkgPath, "pkg-path", "/root/kube1.14.1.tar.gz",
		"Offline installation package storage path")
	kubeCmd.PersistentFlags().StringVar(&command.CobraFlag.ImageRepo, "repo", "k8s.gcr.io",
		"Choose a container registry to pull control plane images from ")
	kubeCmd.PersistentFlags().StringVar(&command.CobraFlag.PodCIDR, "podcidr", "100.64.0.0/10",
		"Specify range of IP addresses for the pod network. If set, the control plane will automatically allocate CIDRs for every node")
	kubeCmd.PersistentFlags().StringVar(&command.CobraFlag.SvcCIDR, "svccidr", "10.96.0.0/12",
		"Use alternative range of IP address for service VIPs.")
	kubeCmd.PersistentFlags().BoolVar(&command.CobraFlag.WithoutCNI, "without-cni", false, "If true we not install cni plugin")
	kubeCmd.PersistentFlags().StringVarP(&command.CobraFlag.K8sVersion, "k8s-version", "v", "v1.14.1",
		"Version is kubernetes version")

	// join kubernentes nodes
	kubeCmd.PersistentFlags().StringSliceVar(&command.CobraFlag.JoinNodeIP, "join-node", []string{},
		"Kubernetes join multi node ip ex. 192.168.0.2-192.168.0.5")
	kubeCmd.PersistentFlags().BoolVar(&command.CobraFlag.ControlPlane, "control-plane", false,
		"Kubernetes join master node,need set --control-plane,Otherwise show join worker node")
}
