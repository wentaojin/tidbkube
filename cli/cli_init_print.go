package cli

import (
	"fmt"

	"github.com/WentaoJin/tidbkube/command"
	"github.com/spf13/cobra"
)

// printCmd define cli program command
var printCmd = &cobra.Command{
	Use:     "print",
	Aliases: []string{"printCmd"},
	Short:   "Print the default configuration file for manual configuration of kubeadm-config.yaml",
	Long:    fmt.Sprintf("%s\nPrint the default configuration file for manual configuration of kubeadm-config.yaml", headerStyle),
	PreRun: func(cmd *cobra.Command, args []string) {
		initProgramRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		switch {
		case command.CobraFlag.TemplateKubeadm:
			command.PrintlnDefaultKubeadmTemplate()
		case command.CobraFlag.TemplateFlannel:
			command.PrintlnDefaultFlannelTemplate()
		case command.CobraFlag.TemplateCalico:
			command.PrintlnDefaultCalicoTemplate()
		default:
			cmd.Help()
		}
	},
}

// register command and flags
func init() {
	initCmd.AddCommand(printCmd)

	printCmd.Flags().BoolVarP(&command.CobraFlag.TemplateKubeadm, "template-kubeadm", "", false,
		" Print Kubeadm-config.yaml default temlpate file")
	printCmd.Flags().BoolVarP(&command.CobraFlag.TemplateCalico, "template-calico", "", false,
		" Print kubernentes network plugin calico default temlpate yaml file")
	printCmd.Flags().BoolVarP(&command.CobraFlag.TemplateFlannel, "template-flannel", "", false,
		" Print kubernentes network plugin flannel default temlpate yaml file")

}
