package sshtun

type TunnelOption func(*TunnelConfig)

// func WithInsecureHostKey() TunnelOption {
// 	return func(tc *TunnelConfig) {
// 		tc.HostKeyCallback = ssh.InsecureIgnoreHostKey()
// 	}
// }

// func WithFixedHostKey(hostKey ssh.PublicKey) TunnelOption {
// 	return func(tc *TunnelConfig) {
// 		tc.HostKeyCallback = ssh.FixedHostKey(hostKey)
// 	}
// }

// func WithCustomHostkeyCallback(callback ssh.HostKeyCallback) TunnelOption {
// 	return func(tc *TunnelConfig) {
// 		tc.HostKeyCallback = callback
// 	}
// }

func WithLogger(logf func(string, ...any)) TunnelOption {
	return func(tc *TunnelConfig) {
		tc.Logf = logf
	}
}
