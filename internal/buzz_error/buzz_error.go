package buzz_error

import "fmt"

type BuzzError struct {
	Code    int
	Message string
}

func (e BuzzError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func InvalidInput(msg string) BuzzError {
	return BuzzError{
		Code:    CodeInvalidInput,
		Message: msg,
	}
}

func InvalidEndpoint(msg string) BuzzError {
	return BuzzError{
		Code:    CodeInvalidEndpoint,
		Message: "Invalid endpoint: " + msg,
	}
}

func HttpError(msg string) BuzzError {
	return BuzzError{
		Code:    CodeHttpError,
		Message: "Http error: " + msg,
	}
}

const (
	CodeSuccess      = 0
	CodeUnknownError = 1

	CodeInvalidInput    = 1001
	CodeInvalidEndpoint = 1002

	CodeHttpError = 2001
)

const (
	MsgOk           = "ok"
	MsgUnknownError = "Unknown error"

	MsgInvalidInput    = "Invalid input"
	MsgInvalidEndpoint = "Invalid endpoint"

	MsgHttpError = "Http error"
)
