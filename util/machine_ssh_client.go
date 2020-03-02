package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// LogicSSHClient struct
type LogicSSHClient struct {
	SSHClient *ssh.Client
}

// CloseSSHClient function
func (l *LogicSSHClient) CloseSSHClient() {
	if l.SSHClient == nil {
		return
	}
	err := l.SSHClient.Close()
	if err != nil {
		log.Fatalln(fmt.Errorf("close LogicSSHClient.SSHClient failed:%s", err.Error()))
		os.Exit(1)
	}
}

// createLogicSSHClient function
func createLogicSSHClient(sshHost, sshPort, sshUser, sshPassword, privateKeyFileName,
	sshKeyPassword string) (client *LogicSSHClient, err error) {
	client = &LogicSSHClient{}
	client.SSHClient, err = createSSHClientAsUser(sshUser, sshPassword, privateKeyFileName, sshKeyPassword, sshHost, sshPort)
	if err != nil {
		return nil, fmt.Errorf("func createSshClientAsUser proxy ssh client for user failed:%s", err.Error())
	}
	return
}

// createSSHClientAsUser function
func createSSHClientAsUser(sshUser, sshPassword, sshKey, sshKeyPassword, sshHost, sshPort string) (client *ssh.Client, err error) {

	targetConfig, errMsg := newSSHClientConfig(sshUser, sshPassword, sshKey, sshKeyPassword)
	if err != nil {
		return nil, fmt.Errorf("func createSshClient proxy ssh config falied:%s", errMsg.Error())
	}
	targetAddr := fmt.Sprintf("%s:%s", sshHost, sshPort)
	client, err = ssh.Dial("tcp", targetAddr, targetConfig)
	if err != nil {
		return nil, fmt.Errorf("func createSshClientAsUser failed to dial:%s", err.Error())
	}
	return

}

func newSSHClientConfig(sshUser, sshPassword, privateKeyFileName, sshKeyPassword string) (config *ssh.ClientConfig, err error) {
	config = &ssh.ClientConfig{
		Timeout:         time.Second * 3,
		User:            sshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if sshPassword != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(sshPassword)}
	} else {
		sshKey := privateKeyFileRead(privateKeyFileName)
		config.Auth = []ssh.AuthMethod{publicKeyAuth([]byte(sshKey), []byte(sshKeyPassword))}
	}
	return
}

func publicKeyAuth(sshKey, keyPassword []byte) ssh.AuthMethod {
	// Create the Signer for this private key.
	var (
		signer ssh.Signer
		err    error
	)
	if len(string(keyPassword)) == 0 {
		signer, err = ssh.ParsePrivateKey(sshKey)
		if err != nil {
			log.Fatalf("Func publicKeyAuth:::Parse ssh key from bytes failed: %s\n.", err.Error())
			return nil
		}
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(sshKey, keyPassword)
		if err != nil {
			log.Fatalf("Func publicKeyAuth:::Parse ssh key from bytes failed: %s\n.", err.Error())
			return nil
		}
	}
	return ssh.PublicKeys(signer)
}

// privateKeyFileRead is used for read public key file,for example: /root/.ssh/id_rsa
func privateKeyFileRead(privateKeyFileName string) string {
	buffer, err := ioutil.ReadFile(privateKeyFileName)
	if err != nil {
		log.Fatalf("Func privateKeyFileRead:::Read public key file failed: %s", err.Error())
		return ""
	}
	return string(buffer)
}
