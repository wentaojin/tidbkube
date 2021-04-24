package command

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/wentaojin/tidbkube/util"
)

// TaskCommandExec function,this function is mainly used to execute a series of only cobra inspect and bootstrap
// subcommand.
func TaskCommandExec(cobraCmd string, cobraFlag FlagCobra, taskStructure interface{}) {
	kubeInstaller := TidbKubeInstaller(cobraFlag)
	// host env check command exec
	switch {
	case cobraFlag.TaskName != "" && cobraFlag.SkipTask == "":
		flagTaskCommandExec(cobraCmd, kubeInstaller, cobraFlag, taskStructure)
	case cobraFlag.TaskName == "" && cobraFlag.SkipTask != "":
		flagSkipTaskCommandExec(cobraCmd, kubeInstaller, cobraFlag, taskStructure)
	case cobraFlag.TaskName != "" && cobraFlag.SkipTask != "":
		flagTaskCommandExec(cobraCmd, kubeInstaller, cobraFlag, taskStructure)
	default:
		fullTaskCommandExec(cobraCmd, kubeInstaller, cobraFlag, taskStructure)
	}
}

// fullTaskCommandExec function,show full task exec
func fullTaskCommandExec(cobraCmd string, kubeInstaller *KubeInstaller, cobraFlag FlagCobra,
	taskStructure interface{}) {
	rt := reflect.TypeOf(taskStructure)
	value := reflect.ValueOf(taskStructure)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		value = value.Elem()
	}
	for i := 0; i < value.NumField(); i++ {
		flagExecTaskCommand(cobraCmd, rt.Field(i).Tag.Get("json"), value.Field(i).String(), kubeInstaller, cobraFlag)
	}
}

// FlagListTaskCommandExec function,show list all cobra subcommands to execute task name and task command
func flagListTaskCommandExec(taskStructure interface{}) {
	log.Printf("Program host environment check task as follow:\n")
	formatStyleWithJSON(taskStructure)
}

// flagSkipTaskCommandExec function,show list skip task cobra subcommands to execute task name and task command
func flagSkipTaskCommandExec(cobraCmd string, kubeInstaller *KubeInstaller, cobraFlag FlagCobra, taskStructure interface{}) {
	rt := reflect.TypeOf(taskStructure)
	value := reflect.ValueOf(taskStructure)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		value = value.Elem()
	}
	var (
		fullTaskSlice map[string]int
		taskSlice     map[int]string
	)
	fullTaskSlice = make(map[string]int)
	taskSlice = make(map[int]string)
	for i := 0; i < value.NumField(); i++ {
		fullTaskSlice[rt.Field(i).Tag.Get("json")] = i
	}
	skipTaskSlice := util.CheckAndSplitCommandArgs(cobraFlag.SkipTask)
	for _, v := range skipTaskSlice {
		if _, ok := fullTaskSlice[v]; ok {
			delete(fullTaskSlice, v)
		}
	}

	// solve go map random access
	sortedKeys := make([]int, 0)
	for task, index := range fullTaskSlice {
		taskSlice[index] = task
		sortedKeys = append(sortedKeys, index)
	}

	sort.Ints(sortedKeys)
	for _, k := range sortedKeys {
		flagExecTaskCommand(cobraCmd, taskSlice[k], value.Field(k).String(), kubeInstaller, cobraFlag)
	}

}

// flagTaskCommandExec function,show list task cobra subcommands to execute task name and task command
func flagTaskCommandExec(cobraCmd string, kubeInstaller *KubeInstaller, cobraFlag FlagCobra,
	taskStructure interface{}) {
	rt := reflect.TypeOf(taskStructure)
	value := reflect.ValueOf(taskStructure)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		value = value.Elem()
	}

	var (
		fullTaskSlice map[string]int
		taskSlice     map[int]string
	)
	fullTaskSlice = make(map[string]int)
	taskSlice = make(map[int]string)
	for i := 0; i < value.NumField(); i++ {
		fullTaskSlice[rt.Field(i).Tag.Get("json")] = i
	}

	taskSliceCmd := util.CheckAndSplitCommandArgs(cobraFlag.TaskName)

	// solve go map random access
	sortedKeys := make([]int, 0)
	for _, v := range taskSliceCmd {
		if _, ok := fullTaskSlice[v]; ok {
			taskSlice[fullTaskSlice[v]] = v
			sortedKeys = append(sortedKeys, fullTaskSlice[v])
		}
	}

	sort.Ints(sortedKeys)
	for _, k := range sortedKeys {
		flagExecTaskCommand(cobraCmd, taskSlice[k], value.Field(k).String(), kubeInstaller, cobraFlag)
	}
}

// execTaskCommand function,basic command entry
func flagExecTaskCommand(cobraCmd string, taskName, taskCommand string, kubeInstaller *KubeInstaller,
	cobraFlag FlagCobra) {
	switch cobraCmd {
	case "inspect":
		TaskInspectCommandSplit(cobraCmd, taskName, taskCommand, kubeInstaller, cobraFlag)
	case "bootstrap":
		TaskBootstrapCommandSplit(cobraCmd, taskName, taskCommand, kubeInstaller, cobraFlag)
	case "init":
		TaskInitCommandSplit(cobraCmd, taskName, taskCommand, kubeInstaller, cobraFlag)
	case "reset":
		TaskResetCommandSplit(cobraCmd, taskName, taskCommand, kubeInstaller, cobraFlag)
	case "join_master":
		TaskJoinMasterCommandSplit(cobraCmd, taskName, taskCommand, kubeInstaller, cobraFlag)
	case "join_worker":
		TaskJoinWorkerCommandSplit(cobraCmd, taskName, taskCommand, kubeInstaller, cobraFlag)
	}

}

func ExecTaskCommandMain(cobraCmd string, taskName string, hostSlice []string, taskCommand string,
	cobraFlag FlagCobra) {
	// ssh private key file password,default null
	sshKeyPassword := ""
	util.SessionCommandExec(cobraCmd, taskName, hostSlice, util.CmdCommand, cobraFlag.SSHUser, cobraFlag.SSHPort,
		cobraFlag.SSHPassword, cobraFlag.SSHPrivateKeyFile, sshKeyPassword, taskCommand, cobraFlag.ScriptFileName,
		cobraFlag.ScriptArg)
}

func SingleHostTaskCommandExec(taskName, sshHost string, taskCommand string, cobraFlag FlagCobra) *util.SSHResultLog {
	var hostSlice []string
	hostSlice = append(hostSlice, sshHost)
	// ssh private key file password,default null
	sshKeyPassword := ""

	log.Printf("Process is running task %s, Please wait a moment.\n", taskName)
	sshResultLog := util.SSHCommandSessionExecMain(taskName, hostSlice, util.CmdCommand, cobraFlag.SSHUser,
		cobraFlag.SSHPort,
		cobraFlag.SSHPassword, cobraFlag.SSHPrivateKeyFile, sshKeyPassword, taskCommand,
		"", "")
	return sshResultLog
}

func ExecSFTPCommandMain(taskName string, hostSlice []string, cobraFlag FlagCobra) {
	// Determine whether a file with the same name exists in the specified host,
	// and compare whether the file with the same name is the same file based on the MD5 value.
	// If the MD5 value is the same, it means that the file exists and is not transmitted through sftp.
	// If the MD5 value is different, then rename the file and then transfer the file.
	sftpHostSlice := remoteHostFileMd5SumCheck(hostSlice, cobraFlag)

	fileDir := filepath.Dir(cobraFlag.PkgPath)
	createDirCmd := fmt.Sprintf(`if [ ! -d %s ]; then mkdir -p %s; fi && echo yes`, fileDir, fileDir)
	checkTaskName := fmt.Sprintf("check_dir_%s_exist", fileDir)
	ExecSSHCommandMain(checkTaskName, sftpHostSlice, createDirCmd, cobraFlag).ResultOutputCheckAndProcessExit(checkTaskName)

	sshKeyPassword := ""
	util.SFTPCommandSessionExecMain(taskName, sftpHostSlice, cobraFlag.SSHPort, cobraFlag.SSHUser, cobraFlag.SSHPassword,
		cobraFlag.SSHPrivateKeyFile, sshKeyPassword, cobraFlag.PkgPath, cobraFlag.PkgPath)
}

func ExecSSHCommandMain(taskName string, hostSlice []string, Cmd string, cobraFlag FlagCobra) *util.SSHResultLog {
	sshKeyPassword := ""
	log.Printf("Process is running task %s, Please wait a moment.\n", taskName)
	sshResultLog := util.SSHCommandSessionExecMain(taskName, hostSlice, util.CmdCommand, cobraFlag.SSHUser,
		cobraFlag.SSHPort,
		cobraFlag.SSHPassword, cobraFlag.SSHPrivateKeyFile, sshKeyPassword, Cmd, "", "")
	return sshResultLog
}

// Host file md5 sum check,determine whether they are the same file according to the md5sum value of the linux command
func remoteHostFileMd5SumCheck(hostSlice []string, cobraFlag FlagCobra) []string {
	md5 := localFileMd5Sum(cobraFlag.PkgPath)
	hostExist, hostNoExist := remoteFileExist(hostSlice, cobraFlag.PkgPath, cobraFlag)
	fileName := filepath.Base(cobraFlag.PkgPath)
	// check exist file remote md5sum,
	// determine whether they are the same file according to the md5sum value of the linux command
	md5Cmd := fmt.Sprintf("md5sum %s | cut -d ' ' -f1", cobraFlag.PkgPath)
	taskName := fmt.Sprintf("check_file_%s_md5", fileName)
	sshResultLog := ExecSSHCommandMain(taskName, hostExist, md5Cmd, cobraFlag)
	sshResultLog.ResultOutputCheckAndProcessExit(taskName)

	diffFileHostSlice := sshResultLog.ResultCheckRemoteHostFileMd5(md5)
	renameFileCmd := fmt.Sprintf("mv %s %s_bak", cobraFlag.PkgPath, cobraFlag.PkgPath)
	taskName = fmt.Sprintf("rename_file_%s_name", fileName)
	ExecSSHCommandMain(taskName, diffFileHostSlice, renameFileCmd, cobraFlag).ResultOutputCheckAndProcessExit(taskName)

	hostNoExist = append(hostNoExist, diffFileHostSlice...)
	return hostNoExist
}

func localFileMd5Sum(pkgFile string) (md5 string) {
	if !util.FileExists(pkgFile) {
		log.Printf("Func sendPackage:::CLI flag pkg-path value file Not Found.")
		os.Exit(1)
	}
	// computer file md5sum
	md5 = util.Md5SumLocalFile(pkgFile)
	return
}

func remoteFileExist(hostSlice []string, remoteFilePath string, cobraFlag FlagCobra) (hostExist, hostNoExist []string) {
	remoteFileName := filepath.Base(remoteFilePath)
	remoteFileCommand := fmt.Sprintf(`if [ ! -f %s ]; then echo "0"; else echo "1" ;fi`, remoteFilePath)

	taskName := fmt.Sprintf("check_file_%s_exist", remoteFileName)
	sshResultLog := ExecSSHCommandMain(taskName, hostSlice, remoteFileCommand, cobraFlag)
	// result log output check whether exist failed hosts
	sshResultLog.ResultOutputCheckAndProcessExit(taskName)
	// remote host file exist result check
	hostExist, hostNoExist = sshResultLog.ResultCheckRemoteHostFileExist()
	return
}

// flag list-task output style with JSON
func formatStyleWithJSON(taskStructure interface{}) {
	rt := reflect.TypeOf(taskStructure)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		log.Fatalf("Check toml config file reflect type error,It's not Struct.")
		os.Exit(1)
	}
	fieldNum := rt.NumField()
	res := make([]string, 0, fieldNum)
	for i := 0; i < fieldNum; i++ {
		tagName := rt.Field(i).Name
		tags := strings.Split(string(rt.Field(i).Tag), "\"")
		if len(tags) > 1 {
			tagName = tags[1]
		}
		res = append(res, tagName)
	}
	js, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		fmt.Printf("Can't read result slice,json formatter failed: %v\n", err)
	}
	fmt.Println(string(js))
}
