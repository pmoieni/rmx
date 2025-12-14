package lib

type DisplayError struct {
	Msg string
}

func (e *DisplayError) Error() string { return e.Msg }
