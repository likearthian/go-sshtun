package sshtun

type ConnState int

const (
	// ConnStateNew represents a new connection.
	ConnStateNew ConnState = iota

	// ConnStateConnected represents a connected connection.
	ConnStateConnected

	// ConnStateDisconnected represents a disconnected connection.
	ConnStateDisconnected

	// ConnStateError represents an errored connection.
	ConnStateError
)

type ConnStateMessage struct {
	State ConnState
	Err   error
	Msg   string
}
