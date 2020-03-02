package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	// CommandFailed show ssh command execute Failed
	CommandFailed = "Failed"
	// CommandSuccess show ssh command execute Success
	CommandSuccess = "Success"
	// CmdCommand show the according to params cmdType value,exec ssh command
	CmdCommand = "command"
	// CmdScript show the according to params cmdType value,exec ssh command
	CmdScript = "script"
	// Cobra* show cobra subcommand  name
	CobraInspect = "inspect"
)

// SSHResultLog struct
type SSHResultLog struct {
	Task           string
	SuccessHosts   []SSHResult
	ErrorHosts     []SSHResult
	TotalHostsInfo string
}

// SSHResult struct,show single machine run result
type SSHResult struct {
	Host      string
	User      string
	Port      string
	Command   string
	Status    string
	Result    string
	StartTime string
	EndTime   string
	CostTime  string
}

// SessionCommandExec function
func SessionCommandExec(cobraCmd string, taskName string, ssHost []string, cmdType string, sshUser, sshPort,
	sshPassword, privateKeyFileName, sshKeyPassword, command,
	scriptFileName, scriptArgs string) {
	log.Printf("Process is running task %s, Please wait a moment.\n", taskName)
	sshResultLog := SSHCommandSessionExecMain(taskName, ssHost, cmdType, sshUser, sshPort,
		sshPassword, privateKeyFileName, sshKeyPassword, command,
		scriptFileName, scriptArgs)
	// result log output check whether exist failed hosts
	sshResultLog.ResultOutputCheckAndProcessExit(taskName)
	// Only cobra subcommand inspect can execute result log success hosts check
	if cobraCmd == CobraInspect {
		sshResultLog.InspectCommandSystemEnvironmentResultCheck(taskName)
	}
}

// SSHCommandSessionExecMain function
func SSHCommandSessionExecMain(taskName string, ssHost []string, cmdType string, sshUser, sshPort, sshPassword,
	privateKeyFileName,
	sshKeyPassword, command string, scriptFileName, scriptArgs string) *SSHResultLog {
	resultLog := &SSHResultLog{}
	pool := NewPool(10, len(ssHost))
	ch := make([]chan SSHResult, len(ssHost))

	// cobra subcommand execution header output
	HeaderColorPrintWithStyle(taskName)
	// cobra subcommand execution content output
	for i, host := range ssHost {
		ch[i] = make(chan SSHResult, 1)
		go func(h string, cmdType string, chr chan SSHResult) {
			pool.Acquire()
			switch cmdType {
			case CmdCommand:
				CommandSSHResult(h, sshUser, sshPort, sshPassword, privateKeyFileName, sshKeyPassword,
					command).CommandSessionExec(chr)
			case CmdScript:
				ScriptSSHResult(h, sshUser, sshPort, sshPassword, privateKeyFileName, sshKeyPassword, scriptFileName,
					scriptArgs).ScriptSessionExec(chr)
			default:
				log.Fatalln("Func SessionCommandExec:::Exec failed because of cmdType value")
			}
			pool.Release()
		}(host, cmdType, ch[i])
		resp := <-ch[i]
		// the results of each host's SSH operation are appended to the result log
		if resp.Status == CommandFailed {
			resultLog.ErrorHosts = append(resultLog.ErrorHosts, resp)
		} else {
			resultLog.SuccessHosts = append(resultLog.SuccessHosts, resp)
		}
		// output the results with style one by one
		FormatResultLogWithBasicStyle(i, resp)
		resultLog.TotalHostsInfo = fmt.Sprintf("%d(Success) + %d(Failed) = %d(Total)", len(resultLog.SuccessHosts),
			len(resultLog.ErrorHosts), len(resultLog.SuccessHosts)+len(resultLog.ErrorHosts))
	}
	pool.Wg.Wait()
	return resultLog
}

// ResultOutputCheckAndProcessExit function,mainly used for ssh session command run,check ssh result whether exist
// error host record.if exist,process exit,otherwise continue run other command.
func (sl *SSHResultLog) ResultOutputCheckAndProcessExit(taskName string) {
	sl.FormatResultLogWithStyle(taskName)
	if len(sl.ErrorHosts) > 0 {
		if taskName == "kubeadm_init_master" {
			log.Fatalf("Program tidbkube run task %s kubernetes install exist error,Please clean and uninstall.\n",
				taskName)
			os.Exit(1)
		}
		log.Fatalf("Program tidbkube run task %s falied,Process exist.\n", taskName)
		os.Exit(1)
	}
}

// InspectCommandSystemEnvironmentResultCheck function
func (sl *SSHResultLog) InspectCommandSystemEnvironmentResultCheck(taskName string) {
	switch taskName {
	case "check_system_version":
		resultCheckSystemVersion(sl)
	case "check_system_cpu":
		resultCheckSystemCPU(sl, 16)
	case "check_system_memory":
		resultCheckSystemMemory(sl, 16)
	case "check_system_network":
		// regular match ping result string percentage, ie packet lose in ping output,
		// if value is equal to 0%,show network normal,not exist packet loss
		resultCheckSystemNetwork(sl, "0%")
	case "check_system_hostname":
		resultCheckSystemHostname(sl)
	case "check_system_mac_addr":
		resultCheckSystemNetworkCardMAC(sl)
	case "check_system_product_id":
		resultCheckSystemProductID(sl)
	case "check_k8s_part_default_port":
		resultCheckK8sPartNode(sl)
	case "check_system_swap":
		resultCheckSystemSwap(sl, 0)
	case "check_docker_install":
		// requireStringValue no show docker not install
		resultCheckDockerInstall(sl, "no")
	case "check_docker_version":
		// require docker version is higher than 19
		resultCheckDockerVersion(sl, "19")
	case "check_ipvsadm_install":
		resultCheckIpvsadmInstall(sl, "no")

	}

}

// FormatResultLogWithStyle function
func (sl *SSHResultLog) FormatResultLogWithStyle(taskName string) {
	sl.Task = taskName
	sl.FormatResultLogWithSimpleStyle()
}

// FormatResultLogWithSimpleStyle function
func (sl *SSHResultLog) FormatResultLogWithSimpleStyle() {
	// cobra subcommand execution footer output
	FooterColorPrintWithStyle(sl.Task, sl.TotalHostsInfo)
	if len(sl.ErrorHosts) > 0 {
		ColorPrintWithTerminalStyle("ERROR", "", "WARNING: ", "Failed hosts, please confirm!\n")
		sl.FormatResultLogWithJSONStyle()
	}
}

func HeaderColorPrintWithStyle(taskName string) {
	ColorPrintWithTerminalStyle("ERROR", "", "*", "", fmt.Sprintf("\n"))
	colorTask := fmt.Sprintf(">>>>>>>>>>>>> Process Running Task %s Start >>>>>>>>>>>>>\n", taskName)
	ColorPrintWithTerminalStyle("INFO", "", colorTask, "")
}

func FooterColorPrintStyleWithoutTotalHostInfo(taskName string) {
	colorTask := fmt.Sprintf(">>>>>>>>>>>>> Process Running Task %s Done >>>>>>>>>>>>>\n", taskName)
	ColorPrintWithTerminalStyle("INFO", "", colorTask, "")
}

func FooterColorPrintWithStyle(taskName, totalHostsInfo string) {
	ColorPrintWithTerminalStyle("INFO", "Total Hosts Running: ", "", fmt.Sprintf("%s\n", totalHostsInfo))
	colorTask := fmt.Sprintf(">>>>>>>>>>>>> Process Running Task %s Done >>>>>>>>>>>>>\n", taskName)
	ColorPrintWithTerminalStyle("INFO", "", colorTask, "")
}

func FormatResultLogWithBasicStyle(k int, sshResult SSHResult) {
	ColorPrintWithTerminalStyle("INFO", "", ">>> ", fmt.Sprintf("NO.%d\n", k+1))
	ColorPrintWithTerminalStyle("INFO", "", "Host: ", fmt.Sprintf("%s\n", sshResult.Host))
	ColorPrintWithTerminalStyle("INFO", "User: ", "", fmt.Sprintf("%s\n", sshResult.User))
	ColorPrintWithTerminalStyle("INFO", "Port: ", "", fmt.Sprintf("%s\n", sshResult.Port))
	ColorPrintWithTerminalStyle("INFO", "Command: ", "", fmt.Sprintf("%s\n", sshResult.Command))
	if sshResult.Status == CommandFailed {
		ColorPrintWithTerminalStyle("ERROR", "Status: ", fmt.Sprintf("%s\n", sshResult.Status))
	} else {
		ColorPrintWithTerminalStyle("INFO", "Status: ", fmt.Sprintf("%s\n", sshResult.Status))
	}
	ColorPrintWithTerminalStyle("INFO", "Result:\n", "", fmt.Sprintf("%s\n", sshResult.Result))
	ColorPrintWithTerminalStyle("INFO", "StartTime: ", "", fmt.Sprintf("%s\n", sshResult.StartTime))
	ColorPrintWithTerminalStyle("INFO", "EndTime: ", "", fmt.Sprintf("%s\n", sshResult.EndTime))
	ColorPrintWithTerminalStyle("INFO", "CostTime: ", "", fmt.Sprintf("%s\n\n", sshResult.CostTime))
}

func (sl *SSHResultLog) FormatResultLogWithJSONStyle() {
	// json.Marshal special html characters are escaped solution
	// json.Marshal default escapeHtml is true, will escape <,>, &
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	err := jsonEncoder.Encode(sl.ErrorHosts)
	//b, err := json.Marshal(logs)
	if err != nil {
		log.Fatalln("Func formatResultToJSON:::JSON Marshal failed: ", err.Error())
		return
	}

	var out bytes.Buffer

	err = json.Indent(&out, []byte(bf.String()), "", "    ")
	if err != nil {
		log.Fatalln("Func formatResultToJSON:::JSON Format failed: ", err.Error())
	}
	out.WriteTo(os.Stdout)

}

// CommandHostInfo struct
type CommandHostInfo struct {
	Host               string
	SSHUser            string
	SSHPort            string
	SSHPassword        string
	PrivateKeyFileName string
	SSHKeyPassword     string
	Command            string
}

// CommandSSHResult function
func CommandSSHResult(sshHost, sshUser, sshPort, sshPassword, privateKeyFileName, sshKeyPassword,
	command string) *CommandHostInfo {
	return &CommandHostInfo{
		Host:               sshHost,
		SSHUser:            sshUser,
		SSHPort:            sshPort,
		SSHPassword:        sshPassword,
		PrivateKeyFileName: privateKeyFileName,
		SSHKeyPassword:     sshKeyPassword,
		Command:            command,
	}
}

// CommandSessionExec function
func (ch *CommandHostInfo) CommandSessionExec(chCmd chan SSHResult) {
	var sshResult SSHResult
	sshResult.Host = ch.Host
	sshResult.User = ch.SSHUser
	sshResult.Port = ch.SSHPort
	sshResult.Command = ch.Command
	startTime := time.Now()
	sshResult.StartTime = startTime.Format("2006-01-02 15:04:05")
	sshSession, err := sshSessionRunCommand(ch.Host, ch.SSHPort, ch.SSHUser, ch.SSHPassword, ch.PrivateKeyFileName,
		ch.SSHKeyPassword,
		ch.Command)
	if err != nil {
		sshResult.Status = CommandFailed
		sshResult.Result = fmt.Sprintf("%s", err)
		endTime := time.Now()
		sshResult.EndTime = endTime.Format("2006-01-02 15:04:05")
		sshResult.CostTime = endTime.Sub(startTime).String()
	} else {
		sshResult.Status = CommandSuccess
		resOut := sshSession.stdOut.String()
		resOut = strings.TrimSpace(resOut)
		sshResult.Result = resOut
		endTime := time.Now()
		sshResult.EndTime = endTime.Format("2006-01-02 15:04:05")
		sshResult.CostTime = endTime.Sub(startTime).String()
	}
	chCmd <- sshResult
}

func sshSessionRunCommand(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
	sshKeyPassword, command string) (*LogicSSHSession, error) {
	sshClient, sshSession, err := createSSHClientToSession(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
		sshKeyPassword)
	if err != nil {
		return nil, err
	}
	defer sshSession.CloseSSHSession()
	defer sshClient.CloseSSHClient()
	err = sshSession.session.Run(command)
	if err != nil {
		resOut := sshSession.stdOut.String()
		resOut = strings.TrimSpace(resOut)
		errMsg := fmt.Sprintf("Func createSSHClientToSession:::Host %s ssh session run shell command failed,"+
			"Command Output: %v,Error: %v",
			sshHost,
			resOut,
			err.Error())
		return nil, errors.New(errMsg)
	}
	if sshSession.stdErr.String() != "" {
		errMsg := fmt.Sprintf("Func createSSHClientToSession:::SSH command %s run failed,Error: %v",
			command, sshSession.stdErr.String())
		return nil, errors.New(errMsg)
	}

	return sshSession, nil
}

// ScriptHostInfo struct
type ScriptHostInfo struct {
	Host               string
	SSHUser            string
	SSHPort            string
	SSHPassword        string
	PrivateKeyFileName string
	SSHKeyPassword     string
	ScriptFileName     string
	ScriptArgs         string
}

func (sh *ScriptHostInfo) ScriptSessionExec(chScript chan SSHResult) {
	var sshResult SSHResult
	sshResult.Host = sh.Host
	sshResult.User = sh.SSHUser
	sshResult.Port = sh.SSHPort
	sshSession, newCmd, err := sshSessionRunScript(sh.Host, sh.SSHPort, sh.SSHUser, sh.SSHPassword,
		sh.PrivateKeyFileName,
		sh.SSHKeyPassword, sh.ScriptFileName, sh.ScriptArgs)
	sshResult.Command = newCmd
	startTime := time.Now()
	sshResult.StartTime = startTime.Format("2006-01-02 15:04:05")
	if err != nil {
		sshResult.Status = CommandFailed
		sshResult.Result = fmt.Sprintf("%s", err)
		endTime := time.Now()
		sshResult.EndTime = endTime.Format("2006-01-02 15:04:05")
		sshResult.CostTime = endTime.Sub(startTime).String()
	} else {
		sshResult.Status = CommandSuccess
		resOut := sshSession.stdOut.String()
		resOut = strings.TrimSpace(resOut)
		sshResult.Result = resOut
		endTime := time.Now()
		sshResult.EndTime = endTime.Format("2006-01-02 15:04:05")
		sshResult.CostTime = endTime.Sub(startTime).String()
	}
	chScript <- sshResult
}

// ScriptSSHResult function
func ScriptSSHResult(host, sshUser, sshPort, sshPassword, privateKeyFileName, sshKeyPassword, scriptFileName, scriptArgs string) *ScriptHostInfo {
	return &ScriptHostInfo{
		Host:               host,
		SSHUser:            sshUser,
		SSHPort:            sshPort,
		SSHPassword:        sshPassword,
		PrivateKeyFileName: privateKeyFileName,
		SSHKeyPassword:     sshKeyPassword,
		ScriptFileName:     scriptFileName,
		ScriptArgs:         scriptArgs,
	}
}

func sshSessionRunScript(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
	sshKeyPassword string, scriptFileName, scriptArgs string) (*LogicSSHSession, string, error) {
	var cmdList []string
	executeScriptCmd := fmt.Sprintf("%s %s %s", "/bin/sh", scriptFileName, scriptArgs)
	cmdList = append(cmdList, executeScriptCmd)
	newCmd := strings.Join(cmdList, " && ")

	sshClient, sshSession, err := createSSHClientToSession(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
		sshKeyPassword)
	if err != nil {
		return nil, newCmd, err
	}
	defer sshSession.CloseSSHSession()
	defer sshClient.CloseSSHClient()

	err = sshSession.session.Run(newCmd)
	if err != nil {
		resOut := sshSession.stdOut.String()
		resOut = strings.TrimSpace(resOut)
		errMsg := fmt.Sprintf("Func createSSHClientToSession:::Host %s ssh session run shell command failed,Command Output: %v,Error: %v",
			sshHost,
			resOut,
			err.Error())
		return nil, newCmd, errors.New(errMsg)
	}
	if sshSession.stdErr.String() != "" {
		errMsg := fmt.Sprintf("Func createSSHClientToSession:::SSH command %s run failed,Error: %v",
			newCmd, sshSession.stdErr.String())
		return nil, newCmd, errors.New(errMsg)
	}
	return sshSession, newCmd, nil
}

func createSSHClientToSession(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
	sshKeyPassword string) (*LogicSSHClient, *LogicSSHSession, error) {
	sshClient, err := createLogicSSHClient(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
		sshKeyPassword)
	if err != nil {
		errMsg := fmt.Sprintf("Func createLogicSSHClient:::Host %s create ssh client failed: %v\n",
			sshHost,
			err.Error())
		return nil, nil, errors.New(errMsg)
	}
	// The closure of the ssh client closure should be in the same function as the ssh session,
	// otherwise the ssh client will close before the ssh session,
	// resulting in failure to read the session data, and an error: use of closed network connection
	// defer sshClient.CloseSSHClient()
	sshSession, err := createLogicSSHSession(sshClient.SSHClient)
	if err != nil {
		errMsg := fmt.Sprintf("Func createLogicSSHSession:::Host %s create ssh session failed: %v\n",
			sshHost,
			err.Error())
		return nil, nil, errors.New(errMsg)
	}
	return sshClient, sshSession, nil
}
