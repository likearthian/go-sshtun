package sshtun

import (
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host            string
	Port            int
	User            string
	Auth            ssh.AuthMethod
	HostKeyCallback ssh.HostKeyCallback
}

// TunnelConfig is the configuration for the tunnel.
type TunnelConfig struct {
	LocalPort   int
	Destination string
	Logf        func(string, ...any)
}

type SSHTunnel struct {
	sshClient *ssh.Client
	tc        TunnelConfig
	stat      chan ConnStateMessage
	logf      func(string, ...any)
}

func defaultLogf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func NewSSHTunnel(conf SSHConfig, localPort int, destination string, options ...TunnelOption) (*SSHTunnel, error) {
	tc := TunnelConfig{
		LocalPort:   localPort,
		Destination: destination,
		Logf:        defaultLogf,
	}

	for _, option := range options {
		option(&tc)
	}

	if conf.Port == 0 {
		conf.Port = 22
	}

	if conf.HostKeyCallback == nil {
		conf.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	sshConfig := &ssh.ClientConfig{
		User: conf.User,
		Auth: []ssh.AuthMethod{
			conf.Auth,
		},
		HostKeyCallback: conf.HostKeyCallback,
		Timeout:         5 * time.Second,
	}

	// Connect to the SSH Server
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conf.Host, conf.Port), sshConfig)
	if err != nil {
		return nil, err
	}

	return &SSHTunnel{
		sshClient: sshClient,
		tc:        tc,
		stat:      make(chan ConnStateMessage),
		logf:      tc.Logf,
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
				t.sendConnMessage(ConnStateMessage{
					State: ConnStateError,
					Err:   err,
				})

				break
			}

			t.sendConnMessage(ConnStateMessage{
				State: ConnStateNew,
				Msg:   fmt.Sprintf("New connection from %s", localConn.RemoteAddr().String()),
			})

			go func() {
				// Establish a connection to the remote server
				remoteConn, err := t.sshClient.Dial("tcp", t.tc.Destination)
				if err != nil {
					t.sendConnMessage(ConnStateMessage{
						State: ConnStateError,
						Err:   err,
					})

					return
				}

				// Start copying data between the local and remote connections
				go copyConn(localConn, remoteConn)
				go copyConn(remoteConn, localConn)
			}()
		}
	}()

	t.sendConnMessage(ConnStateMessage{
		State: ConnStateConnected,
		Msg:   fmt.Sprintf("Tunnel established on localhost:%d", t.tc.LocalPort),
	})

	return nil
}

func (t *SSHTunnel) Close() {
	t.sshClient.Close()
}

func (t *SSHTunnel) ConnState() <-chan ConnStateMessage {
	return t.stat
}

func (t *SSHTunnel) sendConnMessage(msg ConnStateMessage) {
	select {
	case t.stat <- msg:
	default:
		strMsg := msg.Msg
		if msg.Err != nil {
			strMsg = msg.Err.Error()
		}

		t.logf(strMsg)
	}
}

func copyConn(writer, reader net.Conn) {
	defer writer.Close()
	defer reader.Close()
	io.Copy(writer, reader)
}
