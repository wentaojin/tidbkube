package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wentaojin/tidbkube/command"
)

// bootstrapCmd define cli program command
var bootstrapCmd = &cobra.Command{
	Use:     "bootstrap",
	Aliases: []string{"bs"},
	Short:   "Initialize the specified machine or the installation environment of all machines",
	Long:    fmt.Sprintf("%v\nInitialize the specified machine or the installation environment of all machines", headerStyle),
	PreRun: func(cmd *cobra.Command, args []string) {
		initProgramRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		command.BootstrapSystemEnvironment(command.CobraFlag)
	},
}

// register command and flags
func init() {
	rootCmd.AddCommand(bootstrapCmd)
	bootstrapCmd.Flags().StringVarP(&command.CobraFlag.ChronyServer, "chronyServer", "", "pool.ntp.org",
		"specifies that the machine instance under a certain label needs to configure chrony server")
}
