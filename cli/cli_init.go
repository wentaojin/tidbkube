package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wentaojin/tidbkube/command"
)

// initCmd define cli program command
var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"in"},
	Short:   "Initialize the kubernentes HA cluster based on the specified machine",
	Long:    fmt.Sprintf("%v\nInitialize the kubernentes HA cluster based on the specified machine", headerStyle),
	PreRun: func(cmd *cobra.Command, args []string) {
		initProgramRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		command.InitKubernentesCluster(command.CobraFlag)
	},
}

// register command and flags
func init() {
	kubeCmd.AddCommand(initCmd)
}
