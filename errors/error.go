package errors

// ErrorCodeDemuxFailed 错误码
const ErrorCodeDemuxFailed = 0
const ErrorCodeGetIndexFailed = 1

// Error 异常
type Error struct {
	ErrCode int
	ErrMsg  string
}

// NewError 新建异常
func NewError(code int, msg string) *Error {
	return &Error{ErrCode: code, ErrMsg: msg}
}

func (err *Error) Error() string {
	return err.ErrMsg
}
