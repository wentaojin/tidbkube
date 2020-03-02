package cli

import (
	"fmt"

	"github.com/WentaoJin/tidbkube/command"
	"github.com/spf13/cobra"
)

// executeCmd define cli program command,mainly used for manual execute shell command
var executeCmd = &cobra.Command{
	Use:     "execute",
	Aliases: []string{"et"},
	Short:   "Manual execution of shell commands or scripts on all machines or specified machines",
	Long:    fmt.Sprintf("%v\nManual execution of shell commands or scripts on all machines or specified machines", headerStyle),
	PreRun: func(cmd *cobra.Command, args []string) {
		initProgramRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		command.ExecuteCommand(command.CobraFlag)
	},
}

func init() {
	rootCmd.AddCommand(executeCmd)
	executeCmd.Flags().StringVarP(&command.CobraFlag.CommandList, "cmd", "", "",
		"specifies that the machine instance under a certain label needs to run shell command,can run multiple command at once（for example: hostname;date)")
	executeCmd.Flags().StringVarP(&command.CobraFlag.ScriptFileName, "scriptFile", "", "",
		"specifies that the machine instance under a certain label needs to run shell script file")
	executeCmd.Flags().StringVarP(&command.CobraFlag.ScriptArg, "scriptArg", "", "",
		"specifies that the machine instance under a certain label needs to run shell script file args（params）")
}
