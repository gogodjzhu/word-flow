package buzz_error

import "fmt"

type BuzzError struct {
	Code    int
	Message string
}

func (e BuzzError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func InvalidWord(msg string) BuzzError {
	return BuzzError{
		Code:    CodeInvalidWord,
		Message: "Invalid word: " + msg,
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

	CodeInvalidWord     = 1001
	CodeInvalidEndpoint = 1002

	CodeHttpError = 2001
)

const (
	MsgOk           = "ok"
	MsgUnknownError = "Unknown error"

	MsgInvalidWord     = "Invalid word"
	MsgInvalidEndpoint = "Invalid endpoint"

	MsgHttpError = "Http error"
)
