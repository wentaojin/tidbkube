package command

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// FlagCobra show cobra flag variable struct
type FlagCobra struct {
	MasterIP          []string
	WorkerIP          []string
	SSHUser           string
	SSHPort           string
	SSHPassword       string
	SSHPrivateKeyFile string
	TaskName          string
	SkipTask          string
	ListTask          bool
	InitCommandFlag
	JoinCommandFlag
	ResetCommandFlag
	BootstrapCommandFlag
	ExecuteCommandFlag
}

type InitCommandFlag struct {
	KubeadmConfigFile string
	NetworkPluginFile string
	PkgPath           string
	VirtualIP         string
	K8sVersion        string
	ApiServer         string
	ImageRepo         string
	PodCIDR           string
	SvcCIDR           string
	NetworkPlugin     string // network plugin type, calico or flannel etc..
	WithoutCNI        bool   // if true don't install cni plugin,default calico
	// network interface name, like "eth.*|en.*"
	NetworkInterface string
	InitPrintFlag
}

type JoinCommandFlag struct {
	JoinNodeIP   []string
	ControlPlane bool
}

type ResetCommandFlag struct {
	RemoveAll             bool
	RemoveInstallPkg      bool
	RemoveContainerImages bool
	RemoveKubeComponents  bool
}

type InitPrintFlag struct {
	TemplateKubeadm bool
	TemplateCalico  bool
	TemplateFlannel bool
}

type BootstrapCommandFlag struct {
	ChronyServer string
}

type ExecuteCommandFlag struct {
	CommandList    string
	ScriptFileName string
	ScriptArg      string
}

// CobraFlag used by CLI flag variable,
var CobraFlag FlagCobra

// PrintFlagCobraConfig function,cobra flag parameter value print
func (f *FlagCobra) PrintFlagCobraConfig() {
	log.Println("Start print tidbkube config file.")
	y, err := yaml.Marshal(f)
	if err != nil {
		log.Fatalf("Func PrintFlagCobraConfig:::Dump config file failed: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(y))
	log.Println("Print tidbkube config file Done.")
}
