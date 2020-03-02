package cli

import (
	"fmt"

	"github.com/WentaoJin/tidbkube/command"
	"github.com/spf13/cobra"
)

// inspectCmd define cli program command
var inspectCmd = &cobra.Command{
	Use:     "inspect",
	Aliases: []string{"it"},
	Short:   "Check whether the specified machine meet the installation requirements",
	Long:    fmt.Sprintf("%v\nCheck whether the specified machine meet the installation requirements", headerStyle),
	PreRun: func(cmd *cobra.Command, args []string) {
		initProgramRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		command.InspectSystemEnvironment(command.CobraFlag)
	},
}

// register command and flags
func init() {
	rootCmd.AddCommand(inspectCmd)
}
