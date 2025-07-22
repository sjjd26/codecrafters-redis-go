package interfaces

type Command interface {
	// GetType() CommandType
	Handle() (string, error)
}
