package errors

// ErrorCodeDemuxFailed 错误码
const ErrorCodeDemuxFailed = 0

// ErrorCodeGetIndexFailed 错误码索引获取失败
const ErrorCodeGetIndexFailed = 1

// ErrorCodeGetStreamFailed 错误码视频流获取失败
const ErrorCodeGetStreamFailed = 2

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
