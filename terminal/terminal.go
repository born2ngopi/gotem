package terminal

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
	sshTerminal "golang.org/x/crypto/ssh/terminal"
)

func NewTerminal() *terminal {
	sessions := make(map[string]*ssh.Session)
	return &terminal{sessions}
}

func (t terminal) NewSession(config Config) error {

	var (
		conf *ssh.ClientConfig
	)

	if !strings.Contains(config.Host, ":") {
		config.Host = config.Host + ":22"
	}

	if config.AuthorizeKey != "" {

	} else {
		conf = &ssh.ClientConfig{
			User: config.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(config.Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}

	conn, err := ssh.Dial("tcp", config.Host, conf)
	if err != nil {
		return err
	}

	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm-256color", 40, 80, modes); err != nil {
		_ = session.Close()
		return err
	}

	go session.Run("$SHELL || bash")

	t.sessions[config.Host] = session

	return nil
}

func (t *terminal) DeleteSession(host string) error {
	if session, ok := t.sessions[host]; ok {
		session.Close()
		delete(t.sessions, host)
		return nil
	}
	return nil
}

func (t *terminal) Request(host, command string) ([]string, error) {

	if _, ok := t.sessions[host]; !ok {
		return nil, errors.New("host not found")
	}

	var (
		b    bytes.Buffer
		sess = t.sessions[host]
	)
	sess.Stdout = &b
	sess.Stderr = &b

	stdIn, _ := sess.StdinPipe()

	_, err := fmt.Fprintf(stdIn, "%s\n", command)
	if err != nil {
		return nil, err
	}

	if err := sess.Wait(); err != nil {
		return nil, err
	}

	var result []string
	sTerm := sshTerminal.NewTerminal(bufio.NewReadWriter(bufio.NewReader(&b), nil), "< ")
	for {
		line, err := sTerm.ReadLine()
		if err != nil {
			break
		}
		result = append(result, line)
	}

	return result, nil
}
