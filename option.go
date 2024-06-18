package sshtun

import "golang.org/x/crypto/ssh"

type TunnelOption func(*TunnelConfig)

func WithInsecureHostKey() TunnelOption {
	return func(tc *TunnelConfig) {
		tc.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}
}

func WithFixedHostKey(hostKey ssh.PublicKey) TunnelOption {
	return func(tc *TunnelConfig) {
		tc.HostKeyCallback = ssh.FixedHostKey(hostKey)
	}
}

func WithCustomHostkeyCallback(callback ssh.HostKeyCallback) TunnelOption {
	return func(tc *TunnelConfig) {
		tc.HostKeyCallback = callback
	}
}
