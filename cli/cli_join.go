package cli

import (
	"fmt"

	"github.com/wentaojin/tidbkube/command"

	"github.com/spf13/cobra"
)

// joinCmd define cli program command
var joinCmd = &cobra.Command{
	Use:     "join",
	Aliases: []string{"jn"},
	Short:   "Join master Or worker node to the kubernentes HA cluster environment",
	Long:    fmt.Sprintf("%v\nJoin master Or worker node to the kubernentes HA cluster environment", headerStyle),
	PreRun: func(cmd *cobra.Command, args []string) {
		initProgramRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		command.BuildJoinKubernentesNode(command.CobraFlag)
	},
}

func init() {
	kubeCmd.AddCommand(joinCmd)
}
