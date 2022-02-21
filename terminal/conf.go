package terminal

import "golang.org/x/crypto/ssh"

type (
	terminal struct {
		sessions map[string]*ssh.Session
	}

	Config struct {
		AuthorizeKey string
		Host         string
		User         string
		Password     string
	}
)
