package main

import (
	"net/http"
	"os"
	"os/signal"
	"time"

	config "./config"
	hls "./hls"
	logger "./log"
	routers "./routers"
	ts "./ts"
	"github.com/sialot/ezlog"
)

// Log 系统日志
var Log *ezlog.Log

// init 构造方法
func init() {
	config.InitConfig()
	logger.InitLog()
	routers.Init()
	hls.Init()
	ts.Init()
	Log = logger.Log
}

// 未完项
// 索引速度优化、区分文件的索引锁 OK
// 索引计算更新时间，对比ts文件防止索引失效  OK
// 配置文件检测，启动异常处理 OK
// 日志打印 OK
// 其他日常处理 OK

// 服务入口
func main() {

	// 声明路由
	mux := http.NewServeMux()
	mux.HandleFunc("/", routers.Welcome)
	mux.HandleFunc("/hls/", routers.GetMainM3U8)
	mux.HandleFunc("/hls_sub/", routers.GetSubM3U8)
	mux.HandleFunc("/video/", routers.GetVideoStream)

	// 启动服务`
	port, _ := config.Get("server.port")

	svr := http.Server{
		Addr:         ":" + port,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Handler:      mux,
	}

	// 监听退出信号
	quitChan := make(chan os.Signal)
	signal.Notify(quitChan, os.Interrupt, os.Kill)

	// 启动协程，等待信号
	go func() {
		<-quitChan
		logger.FlushLog()
		svr.Close()
		Log.Info("flush log and close server")
	}()

	Log.Info("server started")
	svr.ListenAndServe()
}
