package cli

import (
	"fmt"
	"os"

	"github.com/wentaojin/tidbkube/command"

	"github.com/spf13/cobra"
)

const headerStyle = `
Welcome to
 _____ _ ____  ____  _  __     _          
|_   _(_)  _ \| __ )| |/ /   _| |__   ___ 
  | | | | | | |  _ \| ' / | | | '_ \ / _ \
  | | | | |_| | |_) | . \ |_| | |_) |  __/
  |_| |_|____/|____/|_|\_\__,_|_.__/ \___|
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tidbkube",
	Short: "Program tidbkube is used for deploy k8s with kubeadm and bootstrap tidb environment.",
	Long: fmt.Sprintf("%v\nProgram tidbkube is an application to quickly set up a kubernentes cluster and initialize"+
		" the kubernentes cluster machine environment according to the tidb database requirements", headerStyle),
	PreRun: func(cmd *cobra.Command, args []string) {
		initProgramRun()
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cli *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSliceVarP(&command.CobraFlag.MasterIP, "master", "m", []string{},
		"kubernetes multi-master node ip ex. 192.168.0.2-192.168.0.4")
	rootCmd.PersistentFlags().StringSliceVarP(&command.CobraFlag.WorkerIP, "worker", "w", []string{},
		"kubernetes multi-worker node ip ex. 192.168.0.5-192.168.0.5")
	rootCmd.PersistentFlags().StringVarP(&command.CobraFlag.SSHUser, "user", "u", "root", "server user name for ssh")
	rootCmd.PersistentFlags().StringVarP(&command.CobraFlag.SSHPort, "port", "P", "22", "port for ssh")
	rootCmd.PersistentFlags().StringVarP(&command.CobraFlag.SSHPassword, "password", "p", "", "password for ssh")
	rootCmd.PersistentFlags().StringVarP(&command.CobraFlag.SSHPrivateKeyFile, "private-key", "k", "/root/.ssh/id_rsa", "private key for ssh")
	rootCmd.PersistentFlags().StringVarP(&command.CobraFlag.TaskName, "task", "", "",
		"specifies that the machine needs to run some task (for example: check_system_cpu,check_system_version)")
	rootCmd.PersistentFlags().StringVarP(&command.CobraFlag.SkipTask, "skip-task", "", "",
		"specifies that the machine needs to skip run some task (for example: bootstrap_system_package,bootstrap_chrony_server)")
	rootCmd.PersistentFlags().BoolVarP(&command.CobraFlag.ListTask, "list-task", "", false,
		"specifies that the machine list needs to run task name")

	// Hide some global flags from the execute subcommand
	origHelpFunc := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "execute" || (cmd.Parent() != nil && cmd.Parent().Name() == "execute") {
			cmd.Flags().MarkHidden("task")
			cmd.Flags().MarkHidden("skip-task")
			cmd.Flags().MarkHidden("list-task")
		}
		origHelpFunc(cmd, args)
	})
}

// initProgramRun function,mainly used to json format toml config file and output at the beginning of commands
// running of the CLI program
func initProgramRun() {
	// os stdout header style
	fmt.Printf(headerStyle)
	fmt.Println("")
}
