package ginger

var errMap = make(map[interface{}]string)

type Error interface {
	Code() string
	Error() string
}

func NewError(code string, message string) Error {
	return &errorImp{
		code:    code,
		message: message,
	}
}

func RegisterError(uuid string, message string) {
	errMap[uuid] = message
}

type errorImp struct {
	code    string
	message string
}

func (e *errorImp) Code() string {
	return e.code
}

func (e *errorImp) Error() string {
	return e.message
}
