package translate

type ConnError struct {
	Message string
}

func (ce ConnError) Error() string {
	return ce.Message
}

func (ce ConnError) Is(target error) bool {
	_, ok := target.(ConnError)

	return ok
}
