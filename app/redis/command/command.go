package command

type Command interface {
	// GetType() CommandType
	Execute() (string, error)
}

type WriteCommand interface {
	IsWriteCommand() bool
}

type HandshakeStep int

const (
	HandshakeStepNone HandshakeStep = iota
	HandshakeStepPing
	HandshakeStepReplConfFirst
	HandshakeStepReplConfSecond
	HandshakeStepPsync
)

type HandshakeCommand interface {
	IsHandshakeCommand() bool
	GetHandshakeStep() HandshakeStep
}

type MasterResponseCommand interface {
	IsMasterResponseCommand() bool
}
