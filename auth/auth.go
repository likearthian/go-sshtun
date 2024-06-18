package auth

import (
	"os"

	"golang.org/x/crypto/ssh"
)

func PrivateKeyFile(file string) ssh.AuthMethod {
	buffer, err := os.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(key)
}

func Password(password string) ssh.AuthMethod {
	return ssh.Password(password)
}
