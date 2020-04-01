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
	level, _ := config.SysConfig.Get("log.syslog.level")

	var l = ezlog.LVL_DEBUG
	switch level {
	case "info":
		l = ezlog.LVL_INFO
	case "warn":
		l = ezlog.LVL_WARN
	case "error":
		l = ezlog.LVL_ERROR
	}

	Log = &ezlog.Log{
		Filename:   fileName,
		Pattern:    pattern,
		BufferSize: 0,
		LogLevel:   l}
}

// FlushLog 清空缓冲区
func FlushLog() {
	Log.Flush()
}
