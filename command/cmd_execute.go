package command

import (
	"log"
	"os"

	"github.com/wentaojin/tidbkube/util"
)

// ExecuteCommand function
func ExecuteCommand(cobraFlag FlagCobra) {
	cobraFlag.PrintFlagCobraConfig()
	kubeInstaller := TidbKubeInstaller(cobraFlag)
	taskName := "manual_exec_task"
	executeTaskCommandExec(taskName, kubeInstaller.Hosts, cobraFlag)
}

func executeTaskCommandExec(taskName string, hostSlice []string, cobraFlag FlagCobra) {
	sshKeyPassword := ""
	if cobraFlag.CommandList != "" {
		log.Printf("Process is running task %s, Please wait a moment.\n", taskName)
		util.SessionCommandExec("execute", taskName, hostSlice, util.CmdCommand, cobraFlag.SSHUser, cobraFlag.SSHPort,
			cobraFlag.SSHPassword, cobraFlag.SSHPrivateKeyFile, sshKeyPassword, cobraFlag.CommandList, cobraFlag.ScriptFileName,
			cobraFlag.ScriptArg)
	} else if cobraFlag.ScriptFileName != "" {
		log.Printf("Process is running task %s, Please wait a moment.\n", taskName)
		util.SessionCommandExec("execute", taskName, hostSlice, util.CmdScript, cobraFlag.SSHUser, cobraFlag.SSHPort,
			cobraFlag.SSHPassword, cobraFlag.SSHPrivateKeyFile, sshKeyPassword, cobraFlag.CommandList, cobraFlag.ScriptFileName,
			cobraFlag.ScriptArg)
	} else {
		log.Println("Program tidbkube execute subcommand flag cmd and scriptFile Cannot coexist, Please choose only one.")
		os.Exit(1)
	}

}
