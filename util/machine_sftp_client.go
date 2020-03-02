package util

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/sftp"
)

const oneMBByte = 1024 * 1024

// LogicSFTPClient struct
type LogicSFTPClient struct {
	SFTPClient *sftp.Client
}

// CloseSSHClient function
func (l *LogicSFTPClient) CloseSFTPClient() {
	if l.SFTPClient == nil {
		return
	}
	err := l.SFTPClient.Close()
	if err != nil {
		log.Fatalln(fmt.Errorf("close LogicSFTPClient.SFTPClient failed:%s", err.Error()))
		os.Exit(1)
	}
}

func createLogicSFTPClient(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
	sshKeyPassword string) (sftpClient *LogicSFTPClient, sshClient *LogicSSHClient, err error) {
	sftpClient = &LogicSFTPClient{}
	sshClient, err = createLogicSSHClient(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName, sshKeyPassword)
	if err != nil {
		return nil, nil, err
	}

	// create sftp client
	if sftpClient.SFTPClient, err = sftp.NewClient(sshClient.SSHClient); err != nil {
		return nil, nil, err
	}
	return sftpClient, sshClient, nil
}

// SFTPCopy function,the local file is delivered to other hosts at the same path
func SFTPCopy(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
	sshKeyPassword string, localFilePath, remoteFilePath string) {
	sftpClient, sshClient, err := createLogicSFTPClient(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
		sshKeyPassword)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Func SFTPCopy:::Host %s create logic sftp client failed: %v\n", sshHost, err)
		}
	}()
	if err != nil {
		panic(err)
	}
	defer sshClient.CloseSSHClient()
	defer sftpClient.CloseSFTPClient()

	srcFile, err := os.Open(localFilePath)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Func SFTPCopy:::Host %s open file failed: %v\n", sshHost, err)
		}
	}()
	if err != nil {
		panic(err)
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.SFTPClient.Create(remoteFilePath)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Func SFTPCopy:::Host %s remote file path failed: %v\n", sshHost, err)
		}
	}()
	if err != nil {
		panic(err)
	}
	defer dstFile.Close()

	buf := make([]byte, 100*oneMBByte) // 100MB
	totalMB := 0
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		length, _ := dstFile.Write(buf[0:n])
		totalMB += length / oneMBByte
		log.Printf("Func SFTPCopy:::Host %s transfer file %s total size is: %dMB", sshHost, localFilePath, totalMB)
	}
}
