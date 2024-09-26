package model

type ErrorWithMessage struct {
	Err  error
	Code string
}

func (e ErrorWithMessage) Error() string {
	return e.Err.Error()
}
