package cli

import (
	"fmt"

	"github.com/wentaojin/tidbkube/command"

	"github.com/spf13/cobra"
)

// cleanCmd define cli program command
var cleanCmd = &cobra.Command{
	Use:     "reset",
	Aliases: []string{"rs"},
	Short:   "Reset the kubernentes HA cluster environment,Not include system environment reset(bootstrap)",
	Long: fmt.Sprintf("%v\nClean the kubernentes HA cluster environment,"+
		"Not include system environment clean(bootstrap)", headerStyle),
	PreRun: func(cmd *cobra.Command, args []string) {
		initProgramRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		command.ResetKubernentesCluster(command.CobraFlag)
	},
}

func init() {
	kubeCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVarP(&command.CobraFlag.RemoveAll, "remove", "", false,
		"clean kubernentes cluster env,remove install package、container images and kubernentes binary")
	cleanCmd.Flags().BoolVarP(&command.CobraFlag.RemoveInstallPkg, "remove-install-pkg", "", false,
		"clean kubernentes cluster env,remove install package")
	cleanCmd.Flags().BoolVarP(&command.CobraFlag.RemoveContainerImages, "remove-container-images", "", false,
		"clean kubernentes cluster env,remove container images (docker)")
	cleanCmd.Flags().BoolVarP(&command.CobraFlag.RemoveKubeComponents, "remove-kube-component", "", false,
		"clean kubernentes cluster env,remove kubernentes binary")
}
