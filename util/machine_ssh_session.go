package util

import (
	"bytes"
	"fmt"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Memory area content reading and writing
type safeBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *safeBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}

func (w *safeBuffer) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Bytes()
}

func (w *safeBuffer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
}

func (w *safeBuffer) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.String()
}

// LogicSSHSession struct
type LogicSSHSession struct {
	stdOut  *safeBuffer // ssh session Stdout
	stdErr  *safeBuffer // ssh session StdErr
	session *ssh.Session
}

// CloseSSHSession function
func (s *LogicSSHSession) CloseSSHSession() {
	if s.session != nil {
		s.session.Close()
	}
}

// createLogicSSHSession function,the according ssh.Client create ssh session shell
func createLogicSSHSession(sshClient *ssh.Client) (*LogicSSHSession, error) {
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil,
			fmt.Errorf("SSH client new session failed: %v", err.Error())
	}

	stdOutWriter := new(safeBuffer)
	stdErrWriter := new(safeBuffer)
	sshSession.Stdout = stdOutWriter
	sshSession.Stderr = stdErrWriter

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := sshSession.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, fmt.Errorf("SSH session requestPty failed: %v", err.Error())
	}

	return &LogicSSHSession{
		session: sshSession,
		stdOut:  stdOutWriter,
		stdErr:  stdErrWriter,
	}, nil
}
