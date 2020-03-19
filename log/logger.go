package log

import (
	config "../config"
	"github.com/sialot/ezlog"
)

// Log 日志对象
var Log *ezlog.Log

// InitLog 准备日志
func InitLog() {

	fileName, _ := config.SysConfig.Get("log.syslog.filename")
	pattern, _ := config.SysConfig.Get("log.syslog.pattern")

	Log = &ezlog.Log{
		Filename:   fileName,
		Pattern:    pattern,
		BufferSize: 0}
}

// FlushLog 清空缓冲区
func FlushLog() {
	Log.Flush()
}
