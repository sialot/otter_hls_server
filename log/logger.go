package log

import (
	"github.com/sialot/ezlog"
	config "../config"
)

var Log *ezlog.Log
var MLog *ezlog.Log

// 准备日志
func InitLog() {

	fileName, _ := config.SysConfig.Get("log.syslog.filename")
	pattern, _ := config.SysConfig.Get("log.syslog.pattern")

	Log = &ezlog.Log{
		Filename: fileName,
		Pattern:  pattern}

	fileName, _ = config.SysConfig.Get("log.monitorlog.filename")
	pattern, _ = config.SysConfig.Get("log.monitorlog.pattern")
	MLog = &ezlog.Log{
		Filename:   fileName,
		Pattern:    pattern,
		BufferSize: 100}
	MLog.DisableAutoFlush()
}

// 清空缓冲区
func FlushLog() {
	MLog.Flush()
}
