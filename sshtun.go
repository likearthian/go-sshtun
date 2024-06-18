package sshtun

import (
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClientConfig is the configuration for the SSH client.
type SSHClientConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// SSHServerConfig is the configuration for the SSH server.
type SSHServerConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// TunnelConfig is the configuration for the tunnel.
type TunnelConfig struct {
	LocalPort       int
	Destination     string
	HostKeyCallback ssh.HostKeyCallback
}

type SSHTunnel struct {
	sshClient *ssh.Client
	tc        TunnelConfig
	stat      chan ConnStateMessage
}

func NewSSHTunnel(sshServer string, authMethod ssh.AuthMethod, localPort int, destination string, options ...TunnelOption) (*SSHTunnel, error) {
	tc := TunnelConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		LocalPort:       localPort,
		Destination:     destination,
	}

	for _, option := range options {
		option(&tc)
	}

	sshConfig := &ssh.ClientConfig{
		User: "ssh_user",
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: tc.HostKeyCallback,
		Timeout:         5 * time.Second,
	}

	// Connect to the SSH Server
	sshClient, err := ssh.Dial("tcp", sshServer, sshConfig)
	if err != nil {
		return nil, err
	}

	return &SSHTunnel{
		sshClient: sshClient,
		tc:        tc,
		stat:      make(chan ConnStateMessage),
	}, nil
}

func (t *SSHTunnel) Start() error {
	// Listen on a local network port
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", t.tc.LocalPort))
	if err != nil {
		return err
	}

	go func() {
		defer listener.Close()
		for {
			// Accept a connection
			localConn, err := listener.Accept()
			if err != nil {
				t.stat <- ConnStateMessage{
					State: ConnStateError,
					Err:   err,
				}

				break
			}

			go func() {
				// Establish a connection to the remote server
				remoteConn, err := t.sshClient.Dial("tcp", t.tc.Destination)
				if err != nil {
					t.stat <- ConnStateMessage{
						State: ConnStateError,
						Err:   err,
					}
				}

				// Start copying data between the local and remote connections
				go copyConn(localConn, remoteConn)
				go copyConn(remoteConn, localConn)
			}()
		}
	}()

	t.stat <- ConnStateMessage{
		State: ConnStateConnected,
	}

	return nil
}

func copyConn(writer, reader net.Conn) {
	defer writer.Close()
	defer reader.Close()
	io.Copy(writer, reader)
}
