package util

import "fmt"

// SFTPCommandSessionExecMain function
func SFTPCommandSessionExecMain(taskName string, ssHost []string, sshPort, sshUser, sshPassword, privateKeyFileName,
	sshKeyPassword string, localFilePath, remoteFilePath string) {
	pool := NewPool(10, len(ssHost))

	// cobra subcommand execution header output
	HeaderColorPrintWithStyle(taskName)
	// cobra subcommand execution content output
	for _, host := range ssHost {
		go func(h string) {
			pool.Acquire()
			SFTPCopy(h, sshPort, sshUser, sshPassword, privateKeyFileName, sshKeyPassword, localFilePath, remoteFilePath)
			pool.Release()
		}(host)
	}
	pool.Wg.Wait()
	colorTask := fmt.Sprintf(">>>>>>>>>>>>> Process Running Task %s Done >>>>>>>>>>>>>\n", taskName)
	ColorPrintWithTerminalStyle("INFO", "", colorTask, "")
}
